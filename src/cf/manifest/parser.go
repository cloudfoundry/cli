package manifest

import (
	"generic"
	"io"
	"io/ioutil"
	"launchpad.net/goyaml"
)

func Parse(reader io.Reader) (manifest *Manifest, errs ManifestErrors) {
	yamlBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		errs = append(errs, err)
		return
	}

	yamlMap := generic.NewMap()
	err = goyaml.Unmarshal(yamlBytes, yamlMap)
	if err != nil {
		errs = append(errs, err)
		return
	}

	manifest, errs = NewManifest(yamlMap)
	return
}
