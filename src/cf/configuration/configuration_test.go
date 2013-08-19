package configuration

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestDefaults(t *testing.T){
	config := Default()

	assert.Equal(t, config.Target, "https://api.run.pivotal.io")
	assert.Equal(t, config.ApiVersion, 2)
}
