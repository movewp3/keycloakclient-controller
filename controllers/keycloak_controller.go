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
	"time"

	"github.com/christianwoehrle/keycloakclient-controller/pkg/common"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/christianwoehrle/keycloakclient-controller/version"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	keycloakv1alpha1 "github.com/christianwoehrle/keycloakclient-controller/api/v1alpha1"
)

// KeycloakReconciler reconciles a Keycloak object
type KeycloakReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client   client.Client
	Scheme   *runtime.Scheme
	context  context.Context
	cancel   context.CancelFunc
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=keycloak.org,resources=keycloaks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=keycloak.org,resources=keycloaks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=keycloak.org,resources=keycloaks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Keycloak object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile

// SetupWithManager sets up the controller with the Manager.
func (r *KeycloakReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create a new controller
	c, err := controller.New(KeycloakControllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Keycloak
	err = c.Watch(&source.Kind{Type: &keycloakv1alpha1.Keycloak{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	if err := common.WatchSecondaryResource(c, KeycloakControllerName, common.SecretKind, &corev1.Secret{}, &keycloakv1alpha1.Keycloak{}); err != nil {
		return err
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	//client := mgr.GetClient()
	r.context = ctx
	r.cancel = cancel
	r.recorder = mgr.GetEventRecorderFor(KeycloakControllerName)

	return ctrl.NewControllerManagedBy(mgr).
		For(&keycloakv1alpha1.Keycloak{}).
		Owns(&keycloakv1alpha1.KeycloakRealm{}).
		Complete(r)
}

var logKc = logf.Log.WithName("controller_keycloak")

const (
	KeycloakRequeueDelay      = 150 * time.Second
	KeycloakRequeueDelayError = 60 * time.Second
	KeycloakControllerName    = "keycloak-controller"
)

// newReconciler returns a new reconcile.Reconciler
func NewKeycloakReconciler(mgr manager.Manager) reconcile.Reconciler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	client := mgr.GetClient()

	return &KeycloakReconciler{
		Client:   client,
		Scheme:   mgr.GetScheme(),
		context:  ctx,
		cancel:   cancel,
		recorder: mgr.GetEventRecorderFor(KeycloakControllerName),
	}
}

// blank assignment to verify that KeycloakReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &KeycloakReconciler{}

func (r *KeycloakReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	_ = logf.FromContext(ctx)

	reqLogger := logKc.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Keycloak")

	// Fetch the Keycloak instance
	instance := &keycloakv1alpha1.Keycloak{}

	err := r.Client.Get(r.context, request.NamespacedName, instance)
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
	currentState := common.NewClusterState()

	if instance.Spec.Unmanaged {
		return r.ManageSuccess(instance, currentState)
	}

	if instance.Spec.External.Enabled {
		return r.ManageError(instance, errors.Errorf("if external.enabled is true, unmanaged also needs to be true"))
	}

	// Read current state
	err = currentState.Read(r.context, instance, r.Client)
	if err != nil {
		return r.ManageError(instance, err)
	}

	// Get Action to reconcile current state into desired state

	desiredState := r.ReconcileIt(currentState, instance)

	// Run the actions to reach the desired state
	actionRunner := common.NewClusterActionRunner(r.context, r.Client, r.Scheme, instance)
	err = actionRunner.RunAll(desiredState)
	if err != nil {
		return r.ManageError(instance, err)
	}

	return r.ManageSuccess(instance, currentState)
}

func (r *KeycloakReconciler) ManageError(instance *keycloakv1alpha1.Keycloak, issue error) (reconcile.Result, error) {
	r.recorder.Event(instance, "Warning", "ProcessingError", issue.Error())

	instance.Status.Message = issue.Error()
	instance.Status.Ready = false
	instance.Status.Phase = keycloakv1alpha1.PhaseFailing

	r.setVersion(instance)

	err := r.Client.Status().Update(r.context, instance)
	if err != nil {
		logKc.Error(err, "unable to update status")
	}

	return reconcile.Result{
		RequeueAfter: KeycloakRequeueDelayError,
		Requeue:      true,
	}, nil
}

func (r *KeycloakReconciler) ManageSuccess(instance *keycloakv1alpha1.Keycloak, currentState *common.ClusterState) (reconcile.Result, error) {
	// Check if the resources are ready
	resourcesReady, err := currentState.IsResourcesReady(instance)
	if err != nil {
		return r.ManageError(instance, err)
	}

	instance.Status.Ready = resourcesReady
	instance.Status.Message = ""

	// If resources are ready and we have not errored before now, we are in a reconciling phase
	if resourcesReady {
		instance.Status.Phase = keycloakv1alpha1.PhaseReconciling
	} else {
		instance.Status.Phase = keycloakv1alpha1.PhaseInitialising
	}

	if instance.Spec.External.URL != "" { //nolint
		instance.Status.ExternalURL = instance.Spec.External.URL
	}

	// Let the clients know where the admin credentials are stored
	if currentState.KeycloakAdminSecret != nil {
		instance.Status.CredentialSecret = currentState.KeycloakAdminSecret.Name
	}

	r.setVersion(instance)

	err = r.Client.Status().Update(r.context, instance)
	if err != nil {
		logKc.Error(err, "unable to update status")
		return reconcile.Result{
			RequeueAfter: KeycloakRequeueDelayError,
			Requeue:      true,
		}, nil
	}

	logKc.Info("desired cluster state met")
	return reconcile.Result{RequeueAfter: KeycloakRequeueDelay}, nil
}

func (r *KeycloakReconciler) setVersion(instance *keycloakv1alpha1.Keycloak) {
	instance.Status.Version = version.Version
}
