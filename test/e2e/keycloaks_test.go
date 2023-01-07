package e2e

import (
	"context"
	"fmt"

	keycloakv1alpha1 "github.com/christianwoehrle/keycloakclient-controller/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Keycloak", func() {
	fmt.Println("start Keycloak")

	BeforeEach(func() {
		err := prepareExternalKeycloaksCR()
		Expect(err).To(BeNil())
	})
	AfterEach(func() {
		err := tearDownExternalKeycloaksCR()
		Expect(err).To(BeNil())
	})
	It("keycloak cr is not nil", func() {
		keycloakCR, err := getDeployedKeycloakCR(keycloakNamespace)
		Expect(err).To(BeNil())
		Expect(keycloakCR).NotTo(BeNil())
	})
	It("keycloak status is not nil", func() {
		keycloakCR, err := getDeployedKeycloakCR(keycloakNamespace)
		Expect(err).To(BeNil())
		Expect(keycloakCR.Status).NotTo(BeNil())
	})
	It("keycloak external url is set", func() {
		keycloakCR, err := getDeployedKeycloakCR(keycloakNamespace)
		Expect(err).To(BeNil())
		Expect(keycloakCR.Status.ExternalURL).NotTo(BeEmpty())
	})

	It("keycloak cr is not nil2", func() {

		err := WaitForCondition(getClient(), func(c kubernetes.Interface) error {
			sts, err := getClient().AppsV1().StatefulSets(keycloakNamespace).List(context.TODO(), metav1.ListOptions{})

			if err != nil {
				return errors.Errorf("list StatefulSet failed, ignoring for %v: %v", pollRetryInterval, err)
			}
			if len(sts.Items) == 1 {
				return nil
			}
			return errors.Errorf("should find one Statefulset, as the cluster has been prepared with a keycloak installation")
		})
		Expect(err).To(BeNil())

	})

})

func getKeycloakCR(namespace string) *keycloakv1alpha1.Keycloak {
	return &keycloakv1alpha1.Keycloak{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testKeycloakCRName,
			Namespace: namespace,
			Labels:    CreateLabel(namespace),
		},
		Spec: keycloakv1alpha1.KeycloakSpec{},
	}
}

func getUnmanagedKeycloakCR(namespace string) *keycloakv1alpha1.Keycloak {
	keycloak := getKeycloakCR(namespace)
	keycloak.Name = testKeycloakCRName
	keycloak.Spec.Unmanaged = true
	return keycloak
}

func getExternalKeycloakCR(namespace string, url string) *keycloakv1alpha1.Keycloak {
	keycloak := getUnmanagedKeycloakCR(namespace)
	keycloak.Name = testKeycloakCRName
	keycloak.Labels = CreateLabel(namespace)
	keycloak.Spec.External.Enabled = true
	keycloak.Spec.External.URL = url
	return keycloak
}

func getDeployedKeycloakCR(namespace string) (*keycloakv1alpha1.Keycloak, error) {
	return GetNamespacedKeycloak(namespace, testKeycloakCRName)
}

func getExternalKeycloakSecret(namespace string) (*v1.Secret, error) {
	return getClient().CoreV1().Secrets(namespace).Get(context.TODO(), "credential-"+testKeycloakCRName, metav1.GetOptions{})
}

func prepareUnmanagedKeycloaksCR(namespace string) error {
	keycloakCR := getUnmanagedKeycloakCR(namespace)
	err := CreateKeycloak(keycloakCR)
	if err != nil {
		return err
	}

	err = WaitForKeycloakToBeReady(namespace, testKeycloakCRName)
	if err != nil {
		return err
	}

	return err
}

func prepareExternalKeycloaksCR() error {
	keycloakURL := "http://keycloak.local:80"

	secret, err := getExternalKeycloakSecret(keycloakNamespace)
	if err != nil && !apiErrors.IsNotFound(err) {
		GinkgoWriter.Printf("err: %s\n", err.Error())
		return err
	}
	if err == nil {
		err = DeleteSecret("credential-" + testKeycloakCRName)
		Expect(err).To(BeNil())
	}
	// noe setup secret
	secret = &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "credential-" + testKeycloakCRName,
			Namespace: keycloakNamespace,
		},
		StringData: map[string]string{
			"user":     "admin",
			"password": "admin",
		},
		Type: v1.SecretTypeOpaque,
	}

	err = CreateSecret(secret)
	Expect(err).To(BeNil())
	GinkgoWriter.Printf("secret created\n")

	keycloak, err := GetKeycloak(testKeycloakCRName)
	if err != nil && !apiErrors.IsNotFound(err) {
		return err
	}
	if err == nil {
		err = DeleteKeycloak(keycloak.Name)
		Expect(err).To(BeNil())
	}

	externalKeycloakCR := getExternalKeycloakCR(keycloakNamespace, keycloakURL)

	err = CreateKeycloak(externalKeycloakCR)

	Expect(err).To(BeNil())

	err = WaitForKeycloakToBeReady(keycloakNamespace, testKeycloakCRName)
	if err != nil {
		return err
	}

	return err
}
func tearDownExternalKeycloaksCR() error {
	keycloakURL := "http://keycloak.local:80"

	_, err := getExternalKeycloakSecret(keycloakNamespace)
	GinkgoWriter.Printf("err: %s\n", err.Error())
	if err != nil && !apiErrors.IsNotFound(err) {
		GinkgoWriter.Printf("Secret not found in tearDownExternalKeycloaksCR: %s\n", err.Error())
	}

	err = DeleteSecret("credential-" + testKeycloakCRName)
	if err != nil {
		return err
	}

	GinkgoWriter.Printf("secret deleted\n")

	externalKeycloakCR := getExternalKeycloakCR(keycloakNamespace, keycloakURL)
	GinkgoWriter.Printf("getExternalKeycloakCR\n")

	err = DeleteKeycloak(externalKeycloakCR.Name)

	Expect("err").To(BeNil())
	if err != nil && !apiErrors.IsAlreadyExists(err) {
		return err
	}

	err = WaitForKeycloakToBeReady(keycloakNamespace, testKeycloakCRName)
	if err != nil {
		return err
	}

	return err
}
