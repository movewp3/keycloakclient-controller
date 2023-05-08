package controllers

import (
	keycloakv1alpha1 "github.com/movewp3/keycloakclient-controller/api/v1alpha1"
	"github.com/movewp3/keycloakclient-controller/pkg/common"
	"github.com/movewp3/keycloakclient-controller/pkg/model"
)

type Reconciler interface {
	Reconcile(clusterState *common.ClusterState, cr *keycloakv1alpha1.Keycloak) (common.DesiredClusterState, error)
}

func (i *KeycloakReconciler) ReconcileIt(clusterState *common.ClusterState, cr *keycloakv1alpha1.Keycloak) common.DesiredClusterState {
	desired := common.DesiredClusterState{}

	desired = desired.AddAction(i.GetKeycloakAdminSecretDesiredState(clusterState, cr))

	return desired
}

func (i *KeycloakReconciler) GetKeycloakAdminSecretDesiredState(clusterState *common.ClusterState, cr *keycloakv1alpha1.Keycloak) common.ClusterAction {
	keycloakAdminSecret := model.KeycloakAdminSecret(cr)

	if clusterState.KeycloakAdminSecret == nil {
		return common.GenericCreateAction{
			Ref: keycloakAdminSecret,
			Msg: "Create Keycloak admin secret",
		}
	}
	return common.GenericUpdateAction{
		Ref: model.KeycloakAdminSecretReconciled(cr, clusterState.KeycloakAdminSecret),
		Msg: "Update Keycloak admin secret",
	}
}
