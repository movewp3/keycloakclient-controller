package util

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/movewp3/keycloakclient-controller/pkg/k8sutil"
	"github.com/movewp3/keycloakclient-controller/pkg/model"
	"github.com/pkg/errors"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	config2 "sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type SeedSecretGetter struct {
	secret    string
	lastFetch time.Time
	mutex     sync.Mutex
}

var ssg = &SeedSecretGetter{
	lastFetch: time.Now(),
}

func (g *SeedSecretGetter) GetSecretSeed() string {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// If value is not present or it's been more than 5 minutes, fetch a new value
	if g.secret == "" || time.Since(g.lastFetch) > 5*time.Minute {
		fmt.Println("readtime")
		secret, err := g.readSeedSecret()
		if err == nil {
			g.secret = secret
		}
		g.lastFetch = time.Now()
	}
	return g.secret
}

func (g *SeedSecretGetter) readSeedSecret() (string, error) {
	config, err := config2.GetConfig()
	if err != nil {
		logUtil.Info("cannot get config " + err.Error())
		return "", err
	}

	secretClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		logUtil.Info("cannot get kubernetesClient " + err.Error())
		return "", err
	}

	controllerNamespace, err := k8sutil.GetControllerNamespace()
	if err != nil {
		controllerNamespace = model.DefaultControllerNamespace
	}
	secretSeedSecret, err := secretClient.CoreV1().Secrets(controllerNamespace).Get(context.TODO(), model.SecretSeedSecretName, v12.GetOptions{})
	if err != nil {

		if !kubeerrors.IsNotFound(err) {
			logUtil.Info("error getting secret  " + model.SecretSeedSecretName + " " + err.Error())
		}
		return "", errors.Wrap(err, "failed to get the secretSeed")
	}
	secretSeed := string(secretSeedSecret.Data[model.KeycloakClientSecretSeed])
	if secretSeed == "" {
		return "", errors.Wrap(err, "failed to get the secretSeed")
	}

	return secretSeed, nil
}

var logUtil = logf.Log.WithName("util")

func GetClientShaCode(clientID string) (string, error) {
	secretSeed := ssg.GetSecretSeed()
	if secretSeed == "" {
		logUtil.Info("error getting secret seed " + clientID + " from secretSeed")
	}
	h := sha256.New()
	h.Write([]byte(secretSeed + clientID + model.SALT))
	sha := fmt.Sprintf("%x", h.Sum(nil))
	logUtil.Info("construct Secret for " + clientID + " from secretSeed")
	return sha, nil
}

/*
func getSecretSeed() (string, error) {
	config, err := config2.GetConfig()
	if err != nil {
		logUtil.Info("cannot get config " + err.Error())
		return "", err
	}

	secretClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		logUtil.Info("cannot get kubernetesClient " + err.Error())
		return "", err
	}

	controllerNamespace, err := k8sutil.GetControllerNamespace()
	if err != nil {
		controllerNamespace = model.DefaultControllerNamespace
	}
	secretSeedSecret, err := secretClient.CoreV1().Secrets(controllerNamespace).Get(context.TODO(), model.SecretSeedSecretName, v12.GetOptions{})
	if err != nil {

		if !kubeerrors.IsNotFound(err) {
			logUtil.Info("error getting secret  " + model.SecretSeedSecretName + " " + err.Error())
		}
		return "", errors.Wrap(err, "failed to get the secretSeed")
	}
	secretSeed := string(secretSeedSecret.Data[model.KeycloakClientSecretSeed])
	if secretSeed == "" {
		return "", errors.Wrap(err, "failed to get the secretSeed")
	}

	return secretSeed, nil
}
*/
