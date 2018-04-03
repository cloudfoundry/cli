package director

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

/*
---
name: some-deployment
releases:
- name: release
  version: ver
  stemcell:
    os: ...
    version: ...
*/

type Manifest struct {
	Name string

	Releases []ManifestRelease
}

type ManifestRelease struct {
	Name    string
	Version string

	URL  string
	SHA1 string

	Stemcell ManifestReleaseStemcell
}

type ManifestReleaseStemcell struct {
	OS      string
	Version string
}

func NewManifestFromPath(path string, fs boshsys.FileSystem) (Manifest, error) {
	var manifest Manifest

	bytes, err := fs.ReadFile(path)
	if err != nil {
		return manifest, bosherr.WrapErrorf(err, "Reading manifest '%s'", path)
	}

	err = yaml.Unmarshal(bytes, &manifest)
	if err != nil {
		return manifest, bosherr.WrapError(err, "Unmarshalling manifest")
	}

	return manifest, nil
}

func NewManifestFromBytes(bytes []byte) (Manifest, error) {
	var manifest Manifest

	err := yaml.Unmarshal(bytes, &manifest)
	if err != nil {
		return manifest, bosherr.WrapError(err, "Unmarshalling manifest")
	}

	return manifest, nil
}
