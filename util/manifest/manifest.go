package manifest

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type ManifestCreationError struct {
	Err error
}

func (e ManifestCreationError) Error() string {
	return fmt.Sprintf("Error creating manifest file: %s", e.Err.Error())
}

type Manifest struct {
	Applications []Application `yaml:"applications"`
}

// ReadAndMergeManifests reads the manifest at provided path and returns a
// fully merged set of applications.
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

// WriteApplicationManifest writes the provided application to the given
// filepath. If the filepath does not exist, it will create it.
func WriteApplicationManifest(application Application, filePath string) error {
	manifest := Manifest{Applications: []Application{application}}
	manifestBytes, err := yaml.Marshal(manifest)
	if err != nil {
		return ManifestCreationError{Err: err}
	}

	err = ioutil.WriteFile(filePath, manifestBytes, 0644)
	if err != nil {
		return ManifestCreationError{Err: err}
	}

	return nil

}
