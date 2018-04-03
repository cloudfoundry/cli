package manifest

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

type Manifest struct {
	Name          string   `yaml:"name"`
	Dependencies  []string `yaml:"dependencies"`
	Files         []string `yaml:"files"`
	ExcludedFiles []string `yaml:"excluded_files"`
}

func NewManifestFromPath(path string, fs boshsys.FileSystem) (Manifest, error) {
	var manifest Manifest

	bytes, err := fs.ReadFile(path)
	if err != nil {
		return manifest, bosherr.WrapErrorf(err, "Reading package spec '%s'", path)
	}

	err = yaml.Unmarshal(bytes, &manifest)
	if err != nil {
		return manifest, bosherr.WrapError(err, "Unmarshalling package spec")
	}

	return manifest, nil
}
