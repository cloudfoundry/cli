package manifest

import (
	"generic"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var simpleManifest = `
---
services:
applications:
- name: my-app
`

var manifestWithServices = `
---
services:
- foo-service:
    type: redis-cloud
    provider: amazing-provider
    plan: 200TB
  new-service:
  cool-service:
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

	services := generic.NewMap(manifest.Applications[0].Get("services"))
	assert.Equal(t, services.Count(), 3)
	assert.True(t, services.Has("foo-service"))
	assert.True(t, services.Has("new-service"))
	assert.True(t, services.Has("cool-service"))

	fooServiceFields := generic.NewMap(services.Get("foo-service"))
	assert.Equal(t, fooServiceFields.Get("type").(string), "redis-cloud")
	assert.Equal(t, fooServiceFields.Get("provider").(string), "amazing-provider")
	assert.Equal(t, fooServiceFields.Get("plan").(string), "200TB")
}
