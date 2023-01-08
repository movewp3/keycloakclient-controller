package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/christianwoehrle/keycloakclient-controller/api/v1alpha1"
	keycloakv1alpha1 "github.com/christianwoehrle/keycloakclient-controller/api/v1alpha1"
	"github.com/christianwoehrle/keycloakclient-controller/pkg/client/clientset/versioned"

	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/christianwoehrle/keycloakclient-controller/pkg/common"

	"github.com/pkg/errors"

	//v1alpha1 "github.com/christianwoehrle/keycloakclient-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func getClient() *kubernetes.Clientset {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = filepath.Join(
			os.Getenv("HOME"), ".kube", "config",
		)

	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	Expect(err).NotTo(HaveOccurred())
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		log.Fatal(err)
	}
	return clientset
}

func getKeycloakApiClient() *versioned.Clientset {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = filepath.Join(
			os.Getenv("HOME"), ".kube", "config",
		)

	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	Expect(err).NotTo(HaveOccurred())
	clientset, err := versioned.NewForConfig(config)

	if err != nil {
		log.Fatal(err)
	}
	return clientset
}

type Condition func(c kubernetes.Interface) error

type ResponseCondition func(response *http.Response) error

type ClientCondition func(authenticatedClient common.KeycloakInterface) error

func WaitForCondition(c kubernetes.Interface, cond Condition) error {
	GinkgoWriter.Printf("waiting up to %v for condition", pollTimeout)
	var err error = fmt.Errorf("Cnodition not fulfilled")
	for start := time.Now(); time.Since(start) < pollTimeout; time.Sleep(pollRetryInterval) {
		err = cond(c)
		if err == nil {
			return nil
		}
	}
	return err
}

func WaitForConditionWithClient(keycloakCR keycloakv1alpha1.Keycloak, cond ClientCondition) error {
	return WaitForCondition(getClient(), func(c kubernetes.Interface) error {
		authenticatedClient, err := MakeAuthenticatedClient(keycloakCR)
		if err != nil {
			return err
		}
		return cond(authenticatedClient)
	})
}

func MakeAuthenticatedClient(keycloakCR keycloakv1alpha1.Keycloak) (common.KeycloakInterface, error) {
	keycloakFactory := common.LocalConfigKeycloakFactory{}
	return keycloakFactory.AuthenticatedClient(keycloakCR, true)
}

func WaitForKeycloakToBeReady(namespace string, name string) error {
	return WaitForCondition(getClient(), func(c kubernetes.Interface) error {
		keycloak, err := GetNamespacedKeycloak(namespace, name)
		if err != nil {
			return err
		}

		if !keycloak.Status.Ready {
			GinkgoWriter.Printf("Condition KeycloakReady not yet successful for %s", name)

			keycloakCRParsed, err := json.Marshal(keycloak)
			if err != nil {
				return err
			}

			return errors.Errorf("keycloak is not ready \nCurrent CR value: %s", string(keycloakCRParsed))
		}
		GinkgoWriter.Printf("Condition KeycloakReady successful for %s", name)
		return nil
	})
}

func WaitForRealmToBeReady(namespace string) error {

	return WaitForCondition(getClient(), func(c kubernetes.Interface) error {
		keycloakRealm, err := GetNamespacedKeycloakRealm(namespace, testKeycloakRealmCRName)
		if err != nil {
			return err
		}

		if !keycloakRealm.Status.Ready {
			GinkgoWriter.Printf("Condition RealmReady not yet successful for %s", keycloakRealm.Name)

			keycloakRealmCRParsed, err := json.Marshal(keycloakRealm)
			if err != nil {
				return err
			}

			return errors.Errorf("keycloakRealm is not ready \nCurrent CR value: %s", string(keycloakRealmCRParsed))
		}

		GinkgoWriter.Printf("Condition RealmReady successful %s", keycloakRealm.Name)
		return nil
	})
}

func WaitForClientToBeReady(namespace string, name string) error {

	return WaitForCondition(getClient(), func(c kubernetes.Interface) error {
		keycloakClientCR, err := GetNamespacedKeycloakClient(namespace, name)
		if err != nil {
			return err
		}

		if !keycloakClientCR.Status.Ready {
			GinkgoWriter.Printf("Condition KeycloakClientReady not yet successful for %s", keycloakClientCR.Name)
			keycloakRealmCRParsed, err := json.Marshal(keycloakClientCR)
			if err != nil {
				return err
			}

			return errors.Errorf("keycloakClient is not ready \nCurrent CR value: %s", string(keycloakRealmCRParsed))
		}

		GinkgoWriter.Printf("Condition KeycloakClientReady successful for %s", keycloakClientCR.Name)
		return nil
	})
}

func WaitForClientToBeFailing(namespace string, name string) error {

	return WaitForCondition(getClient(), func(c kubernetes.Interface) error {
		keycloakClientCR, err := GetNamespacedKeycloakClient(namespace, name)
		if err != nil {
			return err
		}

		if keycloakClientCR.Status.Phase != keycloakv1alpha1.PhaseFailing {
			GinkgoWriter.Printf("Condition KeycloakClientFailing not yet successful for %s", keycloakClientCR.Name)
			keycloakRealmCRParsed, err := json.Marshal(keycloakClientCR)

			if err != nil {
				return err
			}

			return errors.Errorf("keycloakClient is not failing \nCurrent CR value: %s", string(keycloakRealmCRParsed))
		}

		GinkgoWriter.Printf("Condition KeycloakClientFailing successful for %s", keycloakClientCR.Name)
		return nil
	})
}

func WaitForResponse(url string, condition ResponseCondition) error {
	return WaitForCondition(getClient(), func(c kubernetes.Interface) error {
		response, err := http.Get(url) //nolint
		if err != nil {
			return err
		}
		defer response.Body.Close()

		err = condition(response)
		if err != nil {
			return err
		}

		return nil
	})
}

func WaitForSuccessResponseToContain(url string, expectedString string) error {
	return WaitForResponse(url, func(response *http.Response) error {
		if response.StatusCode != 200 {
			return errors.Errorf("invalid response from url %s (%v)", url, response.Status)
		}

		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		responseString := string(responseData)

		Expect(responseString).To(ContainSubstring(expectedString))

		return nil
	})
}

func WaitForSuccessResponse(url string) error {
	return WaitForResponse(url, func(response *http.Response) error {
		if response.StatusCode != 200 {
			return errors.Errorf("invalid response from url %s (%v)", url, response.Status)
		}
		return nil
	})
}

func CreateKeycloak(kc *v1alpha1.Keycloak) error {
	_, err := getKeycloakApiClient().KeycloakV1alpha1().Keycloaks(keycloakNamespace).Create(context.Background(), kc, metav1.CreateOptions{})
	return err
}
func CreateKeycloakRealm(kcr *keycloakv1alpha1.KeycloakRealm) error {
	_, err := getKeycloakApiClient().KeycloakV1alpha1().KeycloakRealms(keycloakNamespace).Create(context.Background(), kcr, metav1.CreateOptions{})
	return err
}
func CreateKeycloakClient(kcc *keycloakv1alpha1.KeycloakClient) error {
	_, err := getKeycloakApiClient().KeycloakV1alpha1().KeycloakClients(keycloakNamespace).Create(context.Background(), kcc, metav1.CreateOptions{})
	return err
}
func DeleteSecret(name string) error {
	return getClient().CoreV1().Secrets(keycloakNamespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}
func CreateSecret(secret *v1.Secret) error {
	_, err := getClient().CoreV1().Secrets(keycloakNamespace).Create(context.Background(), secret, metav1.CreateOptions{})
	return err
}

func GetKeycloak(name string) (*keycloakv1alpha1.Keycloak, error) {
	return getKeycloakApiClient().KeycloakV1alpha1().Keycloaks(keycloakNamespace).Get(context.Background(), name, metav1.GetOptions{})
}

func GetNamespacedSecret(namespace string, objectName string, outputObject *v1.Secret) error {
	return getClient().RESTClient().Get().Namespace(namespace).Resource("Secret").Name(objectName).Do(context.Background()).Into(outputObject)
}
func GetNamespacedKeycloak(namespace string, name string) (*keycloakv1alpha1.Keycloak, error) {
	return getKeycloakApiClient().KeycloakV1alpha1().Keycloaks(namespace).Get(context.Background(), name, metav1.GetOptions{})
}
func GetNamespacedKeycloakRealm(namespace string, objectName string) (*keycloakv1alpha1.KeycloakRealm, error) {
	return getKeycloakApiClient().KeycloakV1alpha1().KeycloakRealms(keycloakNamespace).Get(context.Background(), objectName, metav1.GetOptions{})
}

func GetNamespacedKeycloakClient(namespace string, objectName string) (*keycloakv1alpha1.KeycloakClient, error) {
	return getKeycloakApiClient().KeycloakV1alpha1().KeycloakClients(namespace).Get(context.Background(), objectName, metav1.GetOptions{})
}

func UpdateKeycloakClient(namespace string, client *keycloakv1alpha1.KeycloakClient) (*keycloakv1alpha1.KeycloakClient, error) {
	//return getKeycloakApiClient().RESTClient().Post().Resource("keycloakclients").Body(&obj).Do(context.Background()).Into(obj)
	return getKeycloakApiClient().KeycloakV1alpha1().KeycloakClients(namespace).Update(context.Background(), client, metav1.UpdateOptions{})

}
func DeleteKeycloak(name string) error {
	return getKeycloakApiClient().KeycloakV1alpha1().Keycloaks(keycloakNamespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func DeleteKeycloakRealm(name string) error {
	return getKeycloakApiClient().KeycloakV1alpha1().KeycloakRealms(keycloakNamespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func DeleteKeycloakClient(kcc keycloakv1alpha1.KeycloakClient) error {
	return getClient().RESTClient().Delete().Namespace(kcc.Namespace).Resource(kcc.Kind).Body(kcc).Do(context.Background()).Into(&kcc)
}

func CreateLabel(namespace string) map[string]string {
	return map[string]string{"app": "kc-in-" + namespace}
}

func GetSuccessfulResponseBody(url string) ([]byte, error) {
	client := &http.Client{}
	response, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	ret, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
