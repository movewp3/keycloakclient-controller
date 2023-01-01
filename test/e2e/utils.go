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
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/christianwoehrle/keycloakclient-controller/pkg/common"

	"github.com/pkg/errors"

	keycloakv1alpha1 "github.com/christianwoehrle/keycloakclient-controller/api/v1alpha1"
	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func getClient() *kubernetes.Clientset {
	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	err = keycloakv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return clientset

}

type Condition func(t *testing.T, c kubernetes.Interface) error

type ResponseCondition func(response *http.Response) error

type ClientCondition func(authenticatedClient common.KeycloakInterface) error

func WaitForCondition(t *testing.T, c kubernetes.Interface, cond Condition) error {
	t.Logf("waiting up to %v for condition", pollTimeout)
	var err error = fmt.Errorf("Cnodition not fulfilled")
	for start := time.Now(); time.Since(start) < pollTimeout; time.Sleep(pollRetryInterval) {
		err = cond(t, c)
		if err == nil {
			return nil
		}
	}
	return err
}

func WaitForConditionWithClient(t *testing.T, keycloakCR keycloakv1alpha1.Keycloak, cond ClientCondition) error {
	return WaitForCondition(t, getClient(), func(t *testing.T, c kubernetes.Interface) error {
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

// Stolen from https://github.com/kubernetes/kubernetes/blob/master/test/e2e/framework/util.go
// Then rewritten to use internal condition statements.
func WaitForStatefulSetReplicasReady(t *testing.T, c kubernetes.Interface, statefulSetName, ns string) error {
	t.Logf("waiting up to %v for StatefulSet %s to have all replicas ready", pollTimeout, statefulSetName)
	return WaitForCondition(t, c, func(t *testing.T, c kubernetes.Interface) error {
		sts, err := c.AppsV1().StatefulSets(ns).Get(context.TODO(), statefulSetName, metav1.GetOptions{})
		if err != nil {
			return errors.Errorf("get StatefulSet %s failed, ignoring for %v: %v", statefulSetName, pollRetryInterval, err)
		}
		if sts.Status.ReadyReplicas == *sts.Spec.Replicas {
			t.Logf("all %d replicas of StatefulSet %s are ready.", sts.Status.ReadyReplicas, statefulSetName)
			return nil
		}
		return errors.Errorf("statefulSet %s found but there are %d ready replicas and %d total replicas", statefulSetName, sts.Status.ReadyReplicas, *sts.Spec.Replicas)
	})
}

func WaitForKeycloakToBeReady(t *testing.T, namespace string, name string) error {
	keycloakCR := &keycloakv1alpha1.Keycloak{}

	return WaitForCondition(t, getClient(), func(t *testing.T, c kubernetes.Interface) error {
		err := GetNamespacedKeycloak(namespace, name, keycloakCR)
		if err != nil {
			return err
		}

		if !keycloakCR.Status.Ready {
			t.Logf("Condition KeycloakReady not yet successful for %s", keycloakCR.Name)

			keycloakCRParsed, err := json.Marshal(keycloakCR)
			if err != nil {
				return err
			}

			return errors.Errorf("keycloak is not ready \nCurrent CR value: %s", string(keycloakCRParsed))
		}
		t.Logf("Condition KeycloakReady successful for %s", keycloakCR.Name)
		return nil
	})
}

func WaitForRealmToBeReady(t *testing.T, namespace string) error {
	keycloakRealmCR := &keycloakv1alpha1.KeycloakRealm{}

	return WaitForCondition(t, getClient(), func(t *testing.T, c kubernetes.Interface) error {
		err := GetNamespacedKeycloakRealm(namespace, testKeycloakRealmCRName, keycloakRealmCR)
		if err != nil {
			return err
		}

		if !keycloakRealmCR.Status.Ready {
			t.Logf("Condition RealmReady not yet successful for %s", keycloakRealmCR.Name)

			keycloakRealmCRParsed, err := json.Marshal(keycloakRealmCR)
			if err != nil {
				return err
			}

			return errors.Errorf("keycloakRealm is not ready \nCurrent CR value: %s", string(keycloakRealmCRParsed))
		}

		t.Logf("Condition RealmReady successful %s", keycloakRealmCR.Name)
		return nil
	})
}

func WaitForClientToBeReady(t *testing.T, namespace string, name string) error {
	keycloakClientCR := &keycloakv1alpha1.KeycloakClient{}

	return WaitForCondition(t, getClient(), func(t *testing.T, c kubernetes.Interface) error {
		err := GetNamespacedKeycloakClient(namespace, name, keycloakClientCR)
		if err != nil {
			return err
		}

		if !keycloakClientCR.Status.Ready {
			t.Logf("Condition KeycloakClientReady not yet successful for %s", keycloakClientCR.Name)
			keycloakRealmCRParsed, err := json.Marshal(keycloakClientCR)
			if err != nil {
				return err
			}

			return errors.Errorf("keycloakClient is not ready \nCurrent CR value: %s", string(keycloakRealmCRParsed))
		}

		t.Logf("Condition KeycloakClientReady successful for %s", keycloakClientCR.Name)
		return nil
	})
}

func WaitForClientToBeFailing(t *testing.T, namespace string, name string) error {
	keycloakClientCR := &keycloakv1alpha1.KeycloakClient{}

	return WaitForCondition(t, getClient(), func(t *testing.T, c kubernetes.Interface) error {
		err := GetNamespacedKeycloakClient(namespace, name, keycloakClientCR)
		if err != nil {
			return err
		}

		if keycloakClientCR.Status.Phase != keycloakv1alpha1.PhaseFailing {
			t.Logf("Condition KeycloakClientFailing not yet successful for %s", keycloakClientCR.Name)
			keycloakRealmCRParsed, err := json.Marshal(keycloakClientCR)

			if err != nil {
				return err
			}

			return errors.Errorf("keycloakClient is not failing \nCurrent CR value: %s", string(keycloakRealmCRParsed))
		}

		t.Logf("Condition KeycloakClientFailing successful for %s", keycloakClientCR.Name)
		return nil
	})
}

func WaitForResponse(t *testing.T, url string, condition ResponseCondition) error {
	return WaitForCondition(t, getClient(), func(t *testing.T, c kubernetes.Interface) error {
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

func WaitForSuccessResponseToContain(t *testing.T, url string, expectedString string) error {
	return WaitForResponse(t, url, func(response *http.Response) error {
		if response.StatusCode != 200 {
			return errors.Errorf("invalid response from url %s (%v)", url, response.Status)
		}

		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		responseString := string(responseData)

		assert.Contains(t, responseString, expectedString)

		return nil
	})
}

func WaitForSuccessResponse(t *testing.T, url string) error {
	return WaitForResponse(t, url, func(response *http.Response) error {
		if response.StatusCode != 200 {
			return errors.Errorf("invalid response from url %s (%v)", url, response.Status)
		}
		return nil
	})
}

func CreateKeycloak(kc keycloakv1alpha1.Keycloak) error {
	result := keycloakv1alpha1.Keycloak{}
	return getClient().RESTClient().Post().Namespace(kc.Namespace).Resource(kc.Kind).Body(kc).Do(context.Background()).Into(&result)
}
func CreateKeycloakRealm(kcr keycloakv1alpha1.KeycloakRealm) error {
	result := keycloakv1alpha1.KeycloakRealm{}
	return getClient().RESTClient().Post().Namespace(kcr.Namespace).Resource(kcr.Kind).Body(kcr).Do(context.Background()).Into(&result)
}
func CreateKeycloakClient(kcc keycloakv1alpha1.KeycloakClient) error {
	result := keycloakv1alpha1.KeycloakClient{}
	return getClient().RESTClient().Post().Namespace(kcc.Namespace).Resource(kcc.Kind).Body(kcc).Do(context.Background()).Into(&result)
}
func CreateSecret(secret v1.Secret) error {
	result := v1.Secret{}
	return getClient().RESTClient().Post().Namespace(secret.Namespace).Resource(secret.Kind).Body(secret).Do(context.Background()).Into(&result)
}

func GetKeycloak(key dynclient.ObjectKey, kc keycloakv1alpha1.Keycloak) error {
	result := keycloakv1alpha1.Keycloak{}
	return getClient().RESTClient().Post().Namespace(kc.Namespace).Resource(kc.Kind).Body(kc).Do(context.Background()).Into(&result)
}

func GetNamespacedSecret(namespace string, objectName string, outputObject *v1.Secret) error {
	return getClient().RESTClient().Get().Namespace(namespace).Resource("Secret").Name(objectName).Do(context.Background()).Into(outputObject)
}
func GetNamespacedKeycloak(namespace string, objectName string, outputObject *keycloakv1alpha1.Keycloak) error {
	return getClient().RESTClient().Get().Namespace(namespace).Resource("Keycloak").Name(objectName).Do(context.Background()).Into(outputObject)
}
func GetNamespacedKeycloakRealm(namespace string, objectName string, outputObject *keycloakv1alpha1.KeycloakRealm) error {
	return getClient().RESTClient().Get().Namespace(namespace).Resource("KeycloakRealm").Name(objectName).Do(context.Background()).Into(outputObject)
}

func GetNamespacedKeycloakClient(namespace string, objectName string, outputObject *keycloakv1alpha1.KeycloakClient) error {
	return getClient().RESTClient().Get().Namespace(namespace).Resource("KeycloakClient").Name(objectName).Do(context.Background()).Into(outputObject)
}

func UpdateKeycloakClient(obj runtime.Object) error {
	return getClient().RESTClient().Post().Resource("KeycloakClient").Body(obj).Do(context.Background()).Into(obj)
}
func DeleteKeycloak(kc keycloakv1alpha1.KeycloakClient) error {
	return getClient().RESTClient().Delete().Namespace(kc.Namespace).Resource(kc.Kind).Body(kc).Do(context.Background()).Into(&kc)
}

func DeleteKeycloakRealm(kcr keycloakv1alpha1.KeycloakRealm) error {
	return getClient().RESTClient().Delete().Namespace(kcr.Namespace).Resource(kcr.Kind).Body(kcr).Do(context.Background()).Into(&kcr)
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
