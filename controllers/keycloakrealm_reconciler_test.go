package controllers

import (
	"testing"

	v12 "k8s.io/api/core/v1"

	"github.com/movewp3/keycloakclient-controller/api/v1alpha1"
	"github.com/movewp3/keycloakclient-controller/pkg/common"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getDummyRealm() *v1alpha1.KeycloakRealm {
	return &v1alpha1.KeycloakRealm{
		Spec: v1alpha1.KeycloakRealmSpec{
			InstanceSelector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "keycloak",
				},
			},
			Realm: &v1alpha1.KeycloakAPIRealm{
				ID:      "dummy",
				Realm:   "dummy",
				Enabled: true,
			},
		},
	}
}

func getDummyState() *common.RealmState {
	return &common.RealmState{
		Realm:            nil,
		RealmUserSecrets: nil,
		Context:          nil,
		Keycloak:         nil,
	}
}

func TestKeycloakRealmReconciler_Update(t *testing.T) {
	// given
	keycloak := v1alpha1.Keycloak{}
	reconciler := NewDedicatedKeycloakRealmReconciler(keycloak)

	realm := getDummyRealm()
	state := getDummyState()

	// reset user credentials to force the operator to create a password
	state.Realm = realm
	state.RealmUserSecrets = make(map[string]*v12.Secret)

	// when
	desiredState := reconciler.Reconcile(state, realm)

	// then
	// 0 - check keycloak available
	// 1 - no other action added
	assert.IsType(t, &common.PingAction{}, desiredState[0])
	assert.Len(t, desiredState, 1)
}
