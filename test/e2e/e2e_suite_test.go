package e2e_test

import (
	"fmt"
	"testing"

	keycloakv1alpha1 "github.com/movewp3/keycloakclient-controller/api/v1alpha1"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"k8s.io/apimachinery/pkg/runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t,
		"E2e Suite")
	scheme = runtime.NewScheme()

}

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(keycloakv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

var _ = Describe("Keycloak", func() {
	fmt.Println("start")
	It("Test", func() {
		Expect("Test").To(Not(BeEmpty()))
	})

})

var _ = BeforeSuite(func() {
	GinkgoWriter.Println("before")
})
var _ = AfterSuite(func() {
	GinkgoWriter.Println("after")
})
