package manifest

import (
	"io/ioutil"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type Manifest struct {
	Applications []Application `yaml:"applications"`
}

type Application struct {
	DockerImage string `yaml:"-"`
	Name        string `yaml:"name"`
	Path        string `yaml:"path"`
}

func ReadAndMergeManifests(pathToManifest string) ([]Application, error) {
	// Read all manifest files
	raw, err := ioutil.ReadFile(pathToManifest)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	err = yaml.Unmarshal(raw, &manifest)
	if err != nil {
		return nil, err
	}

	for i, app := range manifest.Applications {
		if app.Path != "" && !filepath.IsAbs(app.Path) {
			manifest.Applications[i].Path = filepath.Join(filepath.Dir(pathToManifest), app.Path)
		}
	}

	// Merge all manifest files
	return manifest.Applications, err
}
