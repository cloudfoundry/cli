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
				"name": "bitcoin-miner",
			},
		},
	}))
	assert.NoError(t, err)

	apps := m.Applications
	assert.Equal(t, apps[0].Get("instances"), 3)
	assert.Equal(t, apps[0].Get("memory").(uint64), uint64(512))
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
