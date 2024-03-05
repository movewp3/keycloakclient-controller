package util

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/movewp3/keycloakclient-controller/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestStateManager_Test_(t *testing.T) {
	// given
	clientId := "test1"

	// lastFetch now and secret already know should allow to server from cache whcih works in a unit test
	// without k8s access
	ssg.lastFetch = time.Now()
	ssg.secret = "test"

	secret, err := GetClientShaCode(clientId)

	assert.Nil(t, err)

	h := sha256.New()
	h.Write([]byte(ssg.secret + clientId + model.SALT))
	expecretSha := fmt.Sprintf("%x", h.Sum(nil))

	assert.Equal(t, secret, expecretSha, "Expected a secret")
}
