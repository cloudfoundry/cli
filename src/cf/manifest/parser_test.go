package manifest

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var simpleManifest = `
---
applications:
- name: my-app
`

var manifestWithServices = `
---
services:
- foo-service
- new-service
- cool-service
applications:
- name: db-backed-app
`

func TestParsingApplicationName(t *testing.T) {
	manifest, err := Parse(strings.NewReader(simpleManifest))
	assert.NoError(t, err)
	assert.Equal(t, "my-app", manifest.Applications[0].Get("name").(string))
}

func TestParsingManifestServices(t *testing.T) {
	manifest, err := Parse(strings.NewReader(manifestWithServices))
	assert.NoError(t, err)

	services := manifest.Applications[0].Get("services").([]string)
	assert.Equal(t, len(services), 3)
	assert.Equal(t, services[0], "foo-service")
	assert.Equal(t, services[1], "new-service")
	assert.Equal(t, services[2], "cool-service")
}
