package manifest

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

type Manifest struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`

	CommitHash         string `yaml:"commit_hash"`
	UncommittedChanges bool   `yaml:"uncommitted_changes"`

	Jobs         []JobRef             `yaml:"jobs,omitempty"`
	Packages     []PackageRef         `yaml:"packages,omitempty"`
	CompiledPkgs []CompiledPackageRef `yaml:"compiled_packages,omitempty"`
	License      *LicenseRef          `yaml:"license,omitempty"`
}

type JobRef struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"` // todo deprecate
	Fingerprint string `yaml:"fingerprint"`
	SHA1        string `yaml:"sha1"`
}

type PackageRef struct {
	Name         string   `yaml:"name"`
	Version      string   `yaml:"version"` // todo deprecate
	Fingerprint  string   `yaml:"fingerprint"`
	SHA1         string   `yaml:"sha1"`
	Dependencies []string `yaml:"dependencies"`
}

type CompiledPackageRef struct {
	Name          string   `yaml:"name"`
	Version       string   `yaml:"version"` // todo deprecate
	Fingerprint   string   `yaml:"fingerprint"`
	SHA1          string   `yaml:"sha1"`
	OSVersionSlug string   `yaml:"stemcell"`
	Dependencies  []string `yaml:"dependencies"`
}

type LicenseRef struct {
	Version     string `yaml:"version"` // todo deprecate
	Fingerprint string `yaml:"fingerprint"`
	SHA1        string `yaml:"sha1"`
}

var (
	// Ruby CLI for some reason produces invalid annotations
	invalidBinaryAnnotationReplacer = strings.NewReplacer(
		"sha1: !binary |-", "sha1: !!binary |-",
		"version: !binary |-", "version: !!binary |-",
		"fingerprint: !binary |-", "fingerprint: !!binary |-",
	)
)

func NewManifestFromPath(path string, fs boshsys.FileSystem) (Manifest, error) {
	var manifest Manifest

	bytes, err := fs.ReadFile(path)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Reading manifest '%s'", path)
	}

	str := invalidBinaryAnnotationReplacer.Replace(string(bytes))

	err = yaml.Unmarshal([]byte(str), &manifest)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Parsing release manifest")
	}

	return manifest, nil
}
