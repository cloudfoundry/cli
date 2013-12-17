package manifest

import (
	"fileutils"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

var simple_manifest string

func init() {
	manifest_file, err := os.Open("../../fixtures/manifests/single_app.yml")
	if err != nil {
		println(err.Error())
	}

	simple_manifest = fileutils.ReadFile(manifest_file)
}

func TestParsingApplicationName(t *testing.T) {
	parser := NewManifestParser()
	manifest, err := parser.Parse(strings.NewReader(simple_manifest))

	assert.NoError(t, err)
	assert.Equal(t, "my-app", manifest.Applications[0].Get("name").(string))
}
