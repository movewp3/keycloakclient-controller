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

	"github.com/movewp3/keycloakclient-controller/pkg/common"
	"github.com/pkg/errors"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/client-go/tools/record"

	kc "github.com/movewp3/keycloakclient-controller/api/v1alpha1"
	keycloakv1alpha1 "github.com/movewp3/keycloakclient-controller/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// KeycloakRealmReconciler reconciles a KeycloakRealm object
type KeycloakRealmReconciler struct {
	Client   client.Client
	Scheme   *runtime.Scheme
	context  context.Context
	cancel   context.CancelFunc
	recorder record.EventRecorder
}

const (
	RealmFinalizer         = "realm.cleanup"
	RealmRequeueDelayError = 60 * time.Second
	RealmControllerName    = "controller_keycloakrealm"
)

var logKcr = logf.Log.WithName(RealmControllerName)

// newReconciler returns a new reconcile.Reconciler
func newRealmReconciler(mgr manager.Manager) reconcile.Reconciler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	return &KeycloakRealmReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		cancel:   cancel,
		context:  ctx,
		recorder: mgr.GetEventRecorderFor(RealmControllerName),
	}
}

//+kubebuilder:rbac:groups=keycloak.org,resources=keycloakrealms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=keycloak.org,resources=keycloakrealms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=keycloak.org,resources=keycloakrealms/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KeycloakRealm object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *KeycloakRealmReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)
	reqLogger := logKcr.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KeycloakRealm")

	// Fetch the KeycloakRealm instance
	instance := &kc.KeycloakRealm{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.Spec.Unmanaged {
		return reconcile.Result{Requeue: false}, r.manageSuccess(instance, instance.DeletionTimestamp != nil)
	}

	// If no selector is set we can't figure out which Keycloak instance this realm should
	// be added to. Skip reconcile until a selector has been set.
	if instance.Spec.InstanceSelector == nil {
		logKcr.Info(fmt.Sprintf("realm %v/%v has no instance selector and will be ignored", instance.Namespace, instance.Name))
		return reconcile.Result{Requeue: false}, nil
	}

	keycloaks, err := common.GetMatchingKeycloaks(r.context, r.Client, instance.Spec.InstanceSelector)
	if err != nil {
		return r.ManageError(instance, err)
	}

	logKcr.Info(fmt.Sprintf("found %v matching keycloak(s) for realm %v/%v", len(keycloaks.Items), instance.Namespace, instance.Name))

	// The realm may be applicable to multiple keycloak instances,
	// process all of them
	for _, keycloak := range keycloaks.Items {
		// Get an authenticated keycloak api client for the instance
		keycloakFactory := common.LocalConfigKeycloakFactory{}

		if keycloak.Spec.Unmanaged {
			return r.ManageError(instance, errors.Errorf("realms cannot be created for unmanaged keycloak instances"))
		}

		authenticated, err := keycloakFactory.AuthenticatedClient(keycloak, false)

		if err != nil {
			return r.ManageError(instance, err)
		}

		// Compute the current state of the realm
		realmState := common.NewRealmState(r.context, keycloak)

		logKcr.Info(fmt.Sprintf("read state for keycloak %v/%v, realm %v/%v",
			keycloak.Namespace,
			keycloak.Name,
			instance.Namespace,
			instance.Spec.Realm.Realm))

		err = realmState.Read(instance, authenticated, r.Client)
		if err != nil {
			return r.ManageError(instance, err)
		}

		// Figure out the actions to keep the realms up to date with
		// the desired state
		reconciler := NewDedicatedKeycloakRealmReconciler(keycloak)
		desiredState := reconciler.Reconcile(realmState, instance)
		actionRunner := common.NewClusterAndKeycloakActionRunner(r.context, r.Client, r.Scheme, instance, authenticated)

		// Run all actions to keep the realms updated
		err = actionRunner.RunAll(desiredState)
		if err != nil {
			return r.ManageError(instance, err)
		}
	}

	return reconcile.Result{Requeue: false}, r.manageSuccess(instance, instance.DeletionTimestamp != nil)

}

// SetupWithManager sets up the controller with the Manager.
func (r *KeycloakRealmReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New(RealmControllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KeycloakRealm
	err = c.Watch(source.Kind(mgr.GetCache(), &kc.KeycloakRealm{}), &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	r.context = ctx
	r.cancel = cancel
	r.recorder = mgr.GetEventRecorderFor(RealmControllerName)
	return ctrl.NewControllerManagedBy(mgr).
		For(&keycloakv1alpha1.KeycloakRealm{}).
		Complete(r)
}

// blank assignment to verify that ReconcileKeycloakRealm implements reconcile.Reconciler
var _ reconcile.Reconciler = &KeycloakRealmReconciler{}

func (r *KeycloakRealmReconciler) manageSuccess(realm *kc.KeycloakRealm, deleted bool) error {
	realm.Status.Ready = true
	realm.Status.Message = ""
	realm.Status.Phase = keycloakv1alpha1.PhaseReconciling

	err := r.Client.Status().Update(r.context, realm)
	if err != nil {
		logKcr.Error(err, "unable to update status")
	}

	// Finalizer already set?
	finalizerExists := false
	for _, finalizer := range realm.Finalizers {
		if finalizer == RealmFinalizer {
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
		realm.Finalizers = append(realm.Finalizers, RealmFinalizer)
		logKcr.Info(fmt.Sprintf("added finalizer to keycloak realm %v/%v",
			realm.Namespace,
			realm.Spec.Realm.Realm))

		return r.Client.Update(r.context, realm)
	}

	// Otherwise remove the finalizer
	newFinalizers := []string{}
	for _, finalizer := range realm.Finalizers {
		if finalizer == RealmFinalizer {
			logKcr.Info(fmt.Sprintf("removed finalizer from keycloak realm %v/%v",
				realm.Namespace,
				realm.Spec.Realm.Realm))

			continue
		}
		newFinalizers = append(newFinalizers, finalizer)
	}

	realm.Finalizers = newFinalizers
	return r.Client.Update(r.context, realm)
}

func (r *KeycloakRealmReconciler) ManageError(realm *kc.KeycloakRealm, issue error) (reconcile.Result, error) {
	r.recorder.Event(realm, "Warning", "ProcessingError", issue.Error())

	realm.Status.Message = issue.Error()
	realm.Status.Ready = false
	realm.Status.Phase = keycloakv1alpha1.PhaseFailing

	err := r.Client.Status().Update(r.context, realm)
	if err != nil {
		logKcr.Error(err, "unable to update status")
	}

	return reconcile.Result{
		RequeueAfter: RealmRequeueDelayError,
		Requeue:      true,
	}, nil
}
