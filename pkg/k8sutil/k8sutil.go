package k8sutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// GetWatchNamespace returns the Namespace the operator should be watching for changes
func GetWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", ErrWatchNamespaceEnvVar
	}
	return ns, nil
}

// ErrNoNamespace indicates that a namespace could not be found for the current
// environment
var ErrNoNamespace = fmt.Errorf("namespace not found for current environment")

// ErrRunLocal indicates that the operator is set to run in local mode (this error
// is returned by functions that only work on operators running in cluster mode)
var ErrRunLocal = fmt.Errorf("operator run mode forced to local")

// ErrWatchNamespaceEnvVar indicates that the namespace environment variable is not set
var ErrWatchNamespaceEnvVar = fmt.Errorf("watch namespace env var must be set")

// GetOperatorNamespace returns the namespace the operator should be running in.
func GetControllerNamespace() (string, error) {
	if isRunModeLocal() {
		return "", ErrRunLocal
	}
	nsBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNoNamespace
		}
		return "", err
	}
	ns := strings.TrimSpace(string(nsBytes))
	return ns, nil
}

func isRunModeLocal() bool {
	return !isRunModeCluster()
}

// IsRunInCluster checks if the operator is run in cluster
func isRunModeCluster() bool {
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount")
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}

func isKubeMetaKind(kind string) bool {
	if strings.HasSuffix(kind, "List") ||
		kind == "PatchOptions" ||
		kind == "GetOptions" ||
		kind == "DeleteOptions" ||
		kind == "ExportOptions" ||
		kind == "APIVersions" ||
		kind == "APIGroupList" ||
		kind == "APIResourceList" ||
		kind == "UpdateOptions" ||
		kind == "CreateOptions" ||
		kind == "Status" ||
		kind == "WatchEvent" ||
		kind == "ListOptions" ||
		kind == "APIGroup" {
		return true
	}

	return false
}
