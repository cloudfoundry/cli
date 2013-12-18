package manifest

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var simple_manifest = `
---
applications:
- name: my-app
  memory: 128M
  instances: 3
  host: my-host
`

func TestParsingApplicationName(t *testing.T) {
	manifest, err := Parse(strings.NewReader(simple_manifest))

	assert.Equal(t, "my-app", manifest.Applications[0].Get("name").(string))
	assert.NoError(t, err)
}
