package controllers

import (
	kc "github.com/movewp3/keycloakclient-controller/api/v1alpha1"
	"github.com/movewp3/keycloakclient-controller/pkg/common"
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

	return desired
}

func (i *DedicatedKeycloakRealmReconciler) ReconcileRealmDelete(state *common.RealmState, cr *kc.KeycloakRealm) common.DesiredClusterState {
	desired := common.DesiredClusterState{}
	desired.AddAction(i.getKeycloakDesiredState())
	return desired
}

// Always make sure keycloak is able to respond
func (i *DedicatedKeycloakRealmReconciler) getKeycloakDesiredState() common.ClusterAction {
	return &common.PingAction{
		Msg: "check if keycloak is available",
	}
}
