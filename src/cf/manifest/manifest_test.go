package manifest_test

import (
	"cf/manifest"
	"generic"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestManifestWithGlobalAndAppSpecificProperties(t *testing.T) {
	m, _ := manifest.NewManifest(generic.NewMap(map[string]interface{}{
		"instances": "3",
		"memory":    "512M",
		"applications": []interface{}{
			map[string]interface{}{
				"name": "bitcoin-miner",
			},
		},
	}))

	apps := m.Applications
	assert.Equal(t, apps[0].Get("instances"), 3)
	assert.Equal(t, apps[0].Get("memory").(uint64), uint64(512))
}
