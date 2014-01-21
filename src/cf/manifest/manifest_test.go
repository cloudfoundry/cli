package manifest_test

import (
	"cf/manifest"
	"generic"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestManifestWithGlobalAndAppSpecificProperties(t *testing.T) {
	m, err := manifest.NewManifest(generic.NewMap(map[string]interface{}{
		"instances": "3",
		"memory":    "512M",
		"applications": []interface{}{
			map[string]interface{}{
				"name":     "bitcoin-miner",
				"no-route": true,
			},
		},
	}))
	assert.NoError(t, err)

	apps := m.Applications
	assert.Equal(t, apps[0].Get("instances"), 3)
	assert.Equal(t, apps[0].Get("memory").(uint64), uint64(512))
	assert.True(t, apps[0].Get("no-route").(bool))
}

func TestManifestWithInvalidMemory(t *testing.T) {
	_, err := manifest.NewManifest(generic.NewMap(map[string]interface{}{
		"instances": "3",
		"memory":    "512",
		"applications": []interface{}{
			map[string]interface{}{
				"name": "bitcoin-miner",
			},
		},
	}))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "memory")
}

func TestParsingManifestWithTimeoutSetsHealthCheckTimeout(t *testing.T) {
	m, err := manifest.NewManifest(generic.NewMap(map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"name":    "bitcoin-miner",
				"timeout": "360",
			},
		},
	}))

	assert.NoError(t, err)
	assert.Equal(t, m.Applications[0].Get("health_check_timeout"), 360)
	assert.False(t, m.Applications[0].Has("timeout"))
}

func TestParsingManifestWithEmptyEnvVarIsInvalid(t *testing.T) {
	_, err := manifest.NewManifest(generic.NewMap(map[string]interface{}{
		"env": map[string]interface{}{
			"bar": nil,
		},
		"applications": []interface{}{
			map[string]interface{}{
				"name": "bad app",
			},
		},
	}))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "env var 'bar' should not be null")
}

func TestParsingManifestWithPropertiesReturnsErrors(t *testing.T) {
	_, err := manifest.NewManifest(generic.NewMap(map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"env": map[string]interface{}{
					"bar": "many-${foo}-are-cool",
				},
				"instances": nil,
			},
		},
	}))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Properties are not supported. Found property '${foo}'")
}
