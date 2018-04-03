package manifest

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

type ManifestLock struct {
	Name         string   `yaml:"name"`
	Fingerprint  string   `yaml:"fingerprint"`
	Dependencies []string `yaml:"dependencies,omitempty"`
}

func NewManifestLockFromPath(path string, fs boshsys.FileSystem) (ManifestLock, error) {
	var manifest ManifestLock

	bytes, err := fs.ReadFile(path)
	if err != nil {
		return manifest, bosherr.WrapErrorf(err, "Reading package spec lock '%s'", path)
	}

	err = yaml.Unmarshal(bytes, &manifest)
	if err != nil {
		return manifest, bosherr.WrapError(err, "Unmarshalling package spec lock")
	}

	return manifest, nil
}

func (m ManifestLock) AsBytes() ([]byte, error) {
	if len(m.Name) == 0 {
		return nil, bosherr.Errorf("Expected non-empty package name")
	}

	if len(m.Fingerprint) == 0 {
		return nil, bosherr.Errorf("Expected non-empty package fingerprint")
	}

	return yaml.Marshal(m)
}
