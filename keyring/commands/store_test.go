package commands

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
	"reflect"
	"testing"
)

func TestDoStore(t *testing.T) {
	// Expect an error, because the configuration hasn't been stored yet.
	_, err := keyring.Get(ServiceId, testServerId)
	assert.Error(t, err)
	defer func() {
		err := keyring.Delete(ServiceId, testServerId)
		assert.NoError(t, err)
	}()

	var conf = getTestConf()

	// Store the configuration.
	err = doStore(conf)
	assert.NoError(t, err)

	// Get the stored configuration.
	secret, err := keyring.Get(ServiceId, testServerId)
	assert.NoError(t, err)

	// Validate the stored configuration.
	storedConf := new(storeConfiguration)
	err = json.Unmarshal([]byte(secret), storedConf)
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(conf, *storedConf))
}
