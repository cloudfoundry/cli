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
