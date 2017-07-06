package manifest

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Manifest struct {
	Applications []Application `yaml:"applications"`
}

type Application struct {
	DockerImage string `yaml:"-"`
	Name        string `yaml:"name"`
	Path        string `yaml:"-"`
}

func ReadAndMergeManifests(pathToManifest string) ([]Application, error) {
	// Read all manifest files
	raw, err := ioutil.ReadFile(pathToManifest)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	err = yaml.Unmarshal(raw, &manifest)
	// Merge all manifest files
	// Validate any issues
	return manifest.Applications, err
}
