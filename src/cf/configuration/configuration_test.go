package configuration

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaults(t *testing.T) {
	config := Default()

	assert.Equal(t, config.Target, "https://api.run.pivotal.io")
	assert.Equal(t, config.ApiVersion, "2")
}

func TestSavingAndLoading(t *testing.T) {
	configToSave := Default()
	configToSave.ApiVersion = "3.1.0"
	configToSave.Target = "https://target.example.com"

	configToSave.Save()

	savedConfig, err := Load()

	assert.NoError(t, err)
	assert.Equal(t, savedConfig, configToSave)
}
