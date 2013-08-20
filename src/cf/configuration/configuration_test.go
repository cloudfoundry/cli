package configuration

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaults(t *testing.T) {
	config := Default()

	assert.Equal(t, config.Target, "https://api.run.pivotal.io")
	assert.Equal(t, config.ApiVersion, "2")
	assert.Equal(t, config.AuthorizationEndpoint, "https://login.run.pivotal.io")
	assert.Equal(t, config.AccessToken, "")
}

func TestSavingAndLoading(t *testing.T) {
	configToSave := Default()
	configToSave.ApiVersion = "3.1.0"
	configToSave.Target = "https://api.target.example.com"
	configToSave.AuthorizationEndpoint = "https://login.target.example.com"
	configToSave.AccessToken = "bearer my_access_token"

	configToSave.Save()

	savedConfig, err := Load()

	assert.NoError(t, err)
	assert.Equal(t, savedConfig, configToSave)
}
