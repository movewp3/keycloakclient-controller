package e2e

import (
	"context"
	"fmt"
	"testing"

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
	It("keycloak cr is not nil", func() {
		keycloakCR := getDeployedKeycloakCR(keycloakNamespace)
		Expect(keycloakCR).NotTo(BeNil())
	})
	It("keycloak status is not nil", func() {
		keycloakCR := getDeployedKeycloakCR(keycloakNamespace)
		Expect(keycloakCR.Status).NotTo(BeNil())
	})
	It("keycloak external url is set", func() {
		keycloakCR := getDeployedKeycloakCR(keycloakNamespace)
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

func getDeployedKeycloakCR(namespace string) keycloakv1alpha1.Keycloak {
	keycloakCR := keycloakv1alpha1.Keycloak{}
	_ = GetNamespacedKeycloak(namespace, testKeycloakCRName, &keycloakCR)
	return keycloakCR
}

func getExternalKeycloakSecret(namespace string) (*v1.Secret, error) {
	secret, err := getClient().CoreV1().Secrets(namespace).Get(context.TODO(), "credential-"+testKeycloakCRName, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "credential-" + testKeycloakCRName,
			Namespace: namespace,
		},
		Data:       secret.Data,
		StringData: secret.StringData,
		Type:       secret.Type,
	}, nil
}

func prepareUnmanagedKeycloaksCR(t *testing.T, namespace string) error {
	keycloakCR := getUnmanagedKeycloakCR(namespace)
	err := CreateKeycloak(*keycloakCR)
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
	GinkgoWriter.Printf("err: %s\n", err.Error())
	if err != nil && !apiErrors.IsNotFound(err) {
		return err
	}

	if err != nil && !apiErrors.IsNotFound(err) {
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

		err = CreateSecret(*secret)
		if err != nil {
			return err
		}
	}
	GinkgoWriter.Printf("secret created\n")

	externalKeycloakCR := getExternalKeycloakCR(keycloakNamespace, keycloakURL)
	GinkgoWriter.Printf("getExternalKeycloakCR\n")

	err = CreateKeycloak(*externalKeycloakCR)

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
