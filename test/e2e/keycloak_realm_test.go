package e2e

import (
	"testing"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"

	keycloakv1alpha1 "github.com/christianwoehrle/keycloakclient-controller/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	realmName                  = "test-realm"
	testOperatorIDPDisplayName = "Test Operator IDP"
)

var _ = Describe("KeycloakRealm", func() {

	BeforeEach(func() {
		err := prepareExternalKeycloaksCR()
		Expect(err).To(BeNil())
		prepareKeycloakRealmCR()
	})
	AfterEach(func() {
		err := tearDownExternalKeycloaksRealmCR()
		Expect(err).To(BeNil())
		err = tearDownExternalKeycloaksCR()
		Expect(err).To(BeNil())
	})
	It("keycloakrealm cr is not nil", func() {
		keycloakRealmCR, err := getDeployedKeycloakRealmCR(keycloakNamespace)
		Expect(err).To(BeNil())
		Expect(keycloakRealmCR).NotTo(BeNil())
	})
})

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

func prepareKeycloakRealmCR() error {
	keycloakRealmCR := getKeycloakRealmCR(keycloakNamespace)

	err := CreateKeycloakRealm(keycloakRealmCR)
	if err != nil && !apiErrors.IsAlreadyExists(err) {
		return err
	}
	if err == nil {
		err = DeleteKeycloakRealm(testKeycloakRealmCRName)
	}
	return CreateKeycloakRealm(keycloakRealmCR)

}

func keycloakUnmanagedRealmTest(t *testing.T, namespace string) error {
	keycloakRealmCR := getKeycloakRealmCR(namespace)

	err := CreateKeycloakRealm(keycloakRealmCR)
	if err != nil && !apiErrors.IsAlreadyExists(err) {
		return err
	}

	err = WaitForRealmToBeReady(namespace)
	if err != nil {
		return err
	}

	return err
}

func getDeployedKeycloakRealmCR(namespace string) (*keycloakv1alpha1.KeycloakRealm, error) {
	return GetNamespacedKeycloakRealm(namespace, testKeycloakRealmCRName)
}
func tearDownExternalKeycloaksRealmCR() error {
	err := DeleteKeycloakRealm(testKeycloakRealmCRName)

	Expect(err).To(BeNil())
	return nil
}
