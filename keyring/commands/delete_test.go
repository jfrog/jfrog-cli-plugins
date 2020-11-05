package commands

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoDelete(t *testing.T) {
	// Delete and expect an error, because the config hasn't been stored yet.
	err := doDelete(testServerId)
	assert.Error(t, err)

	var conf = getTestConf()

	// Store the configuration.
	err = doStore(conf)
	assert.NoError(t, err)

	// Delete and expect no error.
	err = doDelete(testServerId)
	assert.NoError(t, err)
}