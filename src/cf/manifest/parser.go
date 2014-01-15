package manifest

import (
	"generic"
	"io"
	"io/ioutil"
	"launchpad.net/goyaml"
)

func Parse(reader io.Reader) (yamlMap generic.Map, err error) {
	yamlBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return
	}

	yamlMap = generic.NewMap()
	err = goyaml.Unmarshal(yamlBytes, yamlMap)
	if err != nil {
		return
	}

	return
}

func ParseToManifest(reader io.Reader) (m *Manifest, errs ManifestErrors) {
	mapp, err := Parse(reader)
	if err != nil {
		errs = append(errs, err)
		return
	}
	return NewManifest(mapp)
}
