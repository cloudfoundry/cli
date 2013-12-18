package manifest

import (
	"generic"
	"io"
	"io/ioutil"
	"launchpad.net/goyaml"
)

func Parse(reader io.Reader) (manifest *Manifest, err error) {
	yamlBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return
	}

	yamlMap := generic.NewEmptyMap()
	err = goyaml.Unmarshal(yamlBytes, yamlMap)
	if err != nil {
		return
	}

	manifest = NewManifest(yamlMap)
	return
}
