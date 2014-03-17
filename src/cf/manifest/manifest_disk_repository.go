package manifest

import (
	"errors"
	"generic"
	"github.com/fraenkel/candiedyaml"
	"io"
	"os"
	"path/filepath"
)

type ManifestRepository interface {
	ReadManifest(string) (manifest *Manifest, errors ManifestErrors)
}

type ManifestDiskRepository struct{}

func NewManifestDiskRepository() (repo ManifestRepository) {
	return ManifestDiskRepository{}
}

func (repo ManifestDiskRepository) ReadManifest(inputPath string) (m *Manifest, errs ManifestErrors) {
	m = NewEmptyManifest()

	manifestPath, err := repo.manifestPath(inputPath)
	if err != nil {
		errs = append(errs, errors.New("Error finding manifest: "+err.Error()))
		return
	}
	m.Path = manifestPath

	mapp, err := repo.readAllYAMLFiles(manifestPath)
	if err != nil {
		errs = append(errs, err)
		return
	}
	m.Data = mapp

	return
}

func (repo ManifestDiskRepository) readAllYAMLFiles(path string) (mergedMap generic.Map, err error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return
	}
	defer file.Close()

	mapp, err := parseManifest(file)
	if err != nil {
		return
	}

	if !mapp.Has("inherit") {
		mergedMap = mapp
		return
	}

	inheritedPath, ok := mapp.Get("inherit").(string)
	if !ok {
		err = errors.New("invalid inherit path in manifest")
		return
	}

	if !filepath.IsAbs(inheritedPath) {
		inheritedPath = filepath.Join(filepath.Dir(path), inheritedPath)
	}

	inheritedMap, err := repo.readAllYAMLFiles(inheritedPath)
	if err != nil {
		return
	}

	mergedMap = generic.DeepMerge(inheritedMap, mapp)
	return
}

func parseManifest(file io.Reader) (yamlMap generic.Map, err error) {
	decoder := candiedyaml.NewDecoder(file)
	yamlMap = generic.NewMap()
	err = decoder.Decode(yamlMap)
	if err != nil {
		return
	}

	if !generic.IsMappable(yamlMap) {
		err = errors.New("Invalid manifest. Expected a map")
		return
	}

	return
}

func (repo ManifestDiskRepository) manifestPath(userSpecifiedPath string) (string, error) {
	fileInfo, err := os.Stat(userSpecifiedPath)
	if err != nil {
		return "", err
	}

	if fileInfo.IsDir() {
		manifestPath := filepath.Join(userSpecifiedPath, "manifest.yml")
		_, err := os.Stat(manifestPath)
		return manifestPath, err
	} else {
		return userSpecifiedPath, nil
	}
}
