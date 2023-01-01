package e2e

import (
	"testing"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"

	keycloakv1alpha1 "github.com/christianwoehrle/keycloakclient-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	realmName                  = "test-realm"
	testOperatorIDPDisplayName = "Test Operator IDP"
)

func NewKeycloakRealmsCRDTestStruct() *CRDTestStruct {
	return &CRDTestStruct{
		testSteps: map[string]deployedOperatorTestStep{
			"unmanagedKeycloakRealmTest": {
				testFunction: keycloakUnmanagedRealmTest,
			},
		},
	}
}

func getKeycloakRealmCR(namespace string) *keycloakv1alpha1.KeycloakRealm {
	return &keycloakv1alpha1.KeycloakRealm{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testKeycloakRealmCRName,
			Namespace: namespace,
			Labels:    CreateLabel(namespace),
		},
		Spec: keycloakv1alpha1.KeycloakRealmSpec{
			InstanceSelector: &metav1.LabelSelector{
				MatchLabels: CreateLabel(namespace),
			},
			Unmanaged: true,
			Realm: &keycloakv1alpha1.KeycloakAPIRealm{
				Enabled: true,
				Realm:   "test-realm",
				ID:      "test-realm",
			},
		},
	}
}

func prepareKeycloakRealmCR(t *testing.T, namespace string) error {
	keycloakRealmCR := getKeycloakRealmCR(namespace)

	err := CreateKeycloakRealm(*keycloakRealmCR)
	if err == nil {
		return nil
	}
	if err != nil && !apiErrors.IsAlreadyExists(err) {
		return err
	}
	if err != nil && apiErrors.IsAlreadyExists(err) {
		err = DeleteKeycloakRealm(*keycloakRealmCR)
		if err != nil {
			return CreateKeycloakRealm(*keycloakRealmCR)
		}
	}
	return nil

}

func keycloakUnmanagedRealmTest(t *testing.T, namespace string) error {
	keycloakRealmCR := getKeycloakRealmCR(namespace)

	err := CreateKeycloakRealm(*keycloakRealmCR)
	if err != nil && !apiErrors.IsAlreadyExists(err) {
		return err
	}

	err = WaitForRealmToBeReady(t, namespace)
	if err != nil {
		return err
	}

	return err
}
