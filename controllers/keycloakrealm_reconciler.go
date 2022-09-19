package controllers

import (
	"fmt"

	kc "github.com/christianwoehrle/keycloakclient-controller/api/v1alpha1"
	"github.com/christianwoehrle/keycloakclient-controller/pkg/common"
)

type RealmReconciler interface {
	Reconcile(cr *kc.KeycloakRealm) error
}

type DedicatedKeycloakRealmReconciler struct { // nolint
	Keycloak kc.Keycloak
}

func NewDedicatedKeycloakRealmReconciler(keycloak kc.Keycloak) *DedicatedKeycloakRealmReconciler {
	return &DedicatedKeycloakRealmReconciler{
		Keycloak: keycloak,
	}
}

func (i *DedicatedKeycloakRealmReconciler) Reconcile(state *common.RealmState, cr *kc.KeycloakRealm) common.DesiredClusterState {
	if cr.DeletionTimestamp == nil {
		return i.ReconcileRealmCreate(state, cr)
	}
	return i.ReconcileRealmDelete(state, cr)
}

func (i *DedicatedKeycloakRealmReconciler) ReconcileRealmCreate(state *common.RealmState, cr *kc.KeycloakRealm) common.DesiredClusterState {
	desired := common.DesiredClusterState{}

	desired.AddAction(i.getKeycloakDesiredState())
	desired.AddAction(i.getDesiredRealmState(state, cr))

	return desired
}

func (i *DedicatedKeycloakRealmReconciler) ReconcileRealmDelete(state *common.RealmState, cr *kc.KeycloakRealm) common.DesiredClusterState {
	desired := common.DesiredClusterState{}
	desired.AddAction(i.getKeycloakDesiredState())
	desired.AddAction(i.getDesiredRealmState(state, cr))
	return desired
}

// Always make sure keycloak is able to respond
func (i *DedicatedKeycloakRealmReconciler) getKeycloakDesiredState() common.ClusterAction {
	return &common.PingAction{
		Msg: "check if keycloak is available",
	}
}

func (i *DedicatedKeycloakRealmReconciler) getDesiredRealmState(state *common.RealmState, cr *kc.KeycloakRealm) common.ClusterAction {
	if cr.DeletionTimestamp != nil {
		return &common.DeleteRealmAction{
			Ref: cr,
			Msg: fmt.Sprintf("removing realm %v/%v", cr.Namespace, cr.Spec.Realm.Realm),
		}
	}

	if state.Realm == nil {
		return &common.CreateRealmAction{
			Ref: cr,
			Msg: fmt.Sprintf("create realm %v/%v", cr.Namespace, cr.Spec.Realm.Realm),
		}
	}

	return nil
}
