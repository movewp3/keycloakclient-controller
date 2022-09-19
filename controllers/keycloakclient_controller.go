/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/christianwoehrle/keycloakclient-controller/pkg/common"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/client-go/tools/record"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/christianwoehrle/keycloakclient-controller/api/v1alpha1"
	kc "github.com/christianwoehrle/keycloakclient-controller/api/v1alpha1"
	keycloakv1alpha1 "github.com/christianwoehrle/keycloakclient-controller/api/v1alpha1"
)

// KeycloakClientReconciler reconciles a KeycloakClient object
type KeycloakClientReconciler struct {
	Client   client.Client
	Scheme   *runtime.Scheme
	context  context.Context
	cancel   context.CancelFunc
	recorder record.EventRecorder
}

var logKcc = logf.Log.WithName("controller_keycloakclient")

const (
	ClientFinalizer         = "client.cleanup"
	ClientRequeueDelayError = 60 * time.Second
	ClientControllerName    = "keycloakclient-controller"
)

//+kubebuilder:rbac:groups=keycloak.org,resources=keycloakclients,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=keycloak.org,resources=keycloakclients/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=keycloak.org,resources=keycloakclients/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KeycloakClient object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *KeycloakClientReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	reqLogger := logKcc.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KeycloakClient")

	// Fetch the KeycloakClient instance
	instance := &kc.KeycloakClient{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	r.adjustCrDefaults(instance)

	// The client may be applicable to multiple keycloak instances,
	// process all of them
	realms, err := common.GetMatchingRealms(r.context, r.Client, instance.Spec.RealmSelector)
	if err != nil {
		return r.ManageError(instance, err)
	}
	logKcc.Info(fmt.Sprintf("found %v matching realm(s) for client %v/%v", len(realms.Items), instance.Namespace, instance.Name))
	for _, realm := range realms.Items {
		keycloaks, err := common.GetMatchingKeycloaks(r.context, r.Client, realm.Spec.InstanceSelector)
		if err != nil {
			return r.ManageError(instance, err)
		}
		logKcc.Info(fmt.Sprintf("found %v matching keycloak(s) for realm %v/%v", len(keycloaks.Items), realm.Namespace, realm.Name))

		for _, keycloak := range keycloaks.Items {
			// Get an authenticated keycloak api client for the instance
			keycloakFactory := common.LocalConfigKeycloakFactory{}
			authenticated, err := keycloakFactory.AuthenticatedClient(keycloak, false)
			if err != nil {
				return r.ManageError(instance, err)
			}

			// Compute the current state of the realm
			logKcc.Info(fmt.Sprintf("got authenticated client for keycloak at %v", authenticated.Endpoint()))
			clientState := common.NewClientState(r.context, realm.DeepCopy(), keycloak)

			logKcc.Info(fmt.Sprintf("read client state for keycloak %v/%v, realm %v/%v, client %v/%v",
				keycloak.Namespace,
				keycloak.Name,
				realm.Namespace,
				realm.Name,
				instance.Namespace,
				instance.Name))

			err = clientState.Read(r.context, instance, authenticated, r.Client)
			if err != nil {
				return r.ManageError(instance, err)
			}

			// Figure out the actions to keep the realms up to date with
			// the desired state
			reconciler := NewDedicatedKeycloakClientReconciler(keycloak)
			desiredState := reconciler.ReconcileIt(clientState, instance)
			actionRunner := common.NewClusterAndKeycloakActionRunner(r.context, r.Client, r.Scheme, instance, authenticated)

			// Run all actions to keep the realms updated
			err = actionRunner.RunAll(desiredState)
			if err != nil {
				return r.ManageError(instance, err)
			}
		}
	}

	return reconcile.Result{Requeue: false}, r.manageSuccess(instance, instance.DeletionTimestamp != nil)

}

// SetupWithManager sets up the controller with the Manager.
func (r *KeycloakClientReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	r.context = ctx
	r.cancel = cancel
	r.recorder = mgr.GetEventRecorderFor(ClientControllerName)

	return ctrl.NewControllerManagedBy(mgr).
		For(&keycloakv1alpha1.KeycloakClient{}).
		Complete(r)
}

// Fills the CR with default values. Nils are not acceptable for Kubernetes.
func (r *KeycloakClientReconciler) adjustCrDefaults(cr *kc.KeycloakClient) {
	if cr.Spec.Client.Attributes == nil {
		cr.Spec.Client.Attributes = make(map[string]string)
	}
	if cr.Spec.Client.Access == nil {
		cr.Spec.Client.Access = make(map[string]bool)
	}
	if cr.Spec.Client.AuthenticationFlowBindingOverrides == nil {
		cr.Spec.Client.AuthenticationFlowBindingOverrides = make(map[string]string)
	}
}

func (r *KeycloakClientReconciler) manageSuccess(client *kc.KeycloakClient, deleted bool) error {
	client.Status.Ready = true
	client.Status.Message = ""
	client.Status.Phase = v1alpha1.PhaseReconciling

	err := r.Client.Status().Update(r.context, client)
	if err != nil {
		logKcc.Error(err, "unable to update status")
	}

	// Finalizer already set?
	finalizerExists := false
	for _, finalizer := range client.Finalizers {
		if finalizer == ClientFinalizer {
			finalizerExists = true
			break
		}
	}

	// Resource created and finalizer exists: nothing to do
	if !deleted && finalizerExists {
		return nil
	}

	// Resource created and finalizer does not exist: add finalizer
	if !deleted && !finalizerExists {
		client.Finalizers = append(client.Finalizers, ClientFinalizer)
		logKcc.Info(fmt.Sprintf("added finalizer to keycloak client %v/%v",
			client.Namespace,
			client.Spec.Client.ClientID))

		return r.Client.Update(r.context, client)
	}

	// Otherwise remove the finalizer
	newFinalizers := []string{}
	for _, finalizer := range client.Finalizers {
		if finalizer == ClientFinalizer {
			logKcc.Info(fmt.Sprintf("removed finalizer from keycloak client %v/%v",
				client.Namespace,
				client.Spec.Client.ClientID))

			continue
		}
		newFinalizers = append(newFinalizers, finalizer)
	}

	client.Finalizers = newFinalizers
	return r.Client.Update(r.context, client)
}

func (r *KeycloakClientReconciler) ManageError(realm *kc.KeycloakClient, issue error) (reconcile.Result, error) {
	r.recorder.Event(realm, "Warning", "ProcessingError", issue.Error())

	realm.Status.Message = issue.Error()
	realm.Status.Ready = false
	realm.Status.Phase = v1alpha1.PhaseFailing

	err := r.Client.Status().Update(r.context, realm)
	if err != nil {
		logKcc.Error(err, "unable to update status")
	}

	return reconcile.Result{
		RequeueAfter: ClientRequeueDelayError,
		Requeue:      true,
	}, nil
}
