package manifest

import (
	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/generic"
	"io"
	"os"
	"path/filepath"
)

type ManifestRepository interface {
	ReadManifest(string) (*Manifest, error)
}

type ManifestDiskRepository struct{}

func NewManifestDiskRepository() (repo ManifestRepository) {
	return ManifestDiskRepository{}
}

func (repo ManifestDiskRepository) ReadManifest(inputPath string) (*Manifest, error) {
	m := NewEmptyManifest()
	manifestPath, err := repo.manifestPath(inputPath)

	if err != nil {
		return m, errors.NewWithError(T("Error finding manifest"), err)
	}

	m.Path = manifestPath

	mapp, err := repo.readAllYAMLFiles(manifestPath)
	if err != nil {
		return m, err
	}

	m.Data = mapp

	return m, nil
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
		err = errors.New(T("invalid inherit path in manifest"))
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
		err = errors.New(T("Invalid manifest. Expected a map"))
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
		manifestPaths := []string{
			filepath.Join(userSpecifiedPath, "manifest.yml"),
			filepath.Join(userSpecifiedPath, "manifest.yaml"),
		}
		var err error
		for _, manifestPath := range manifestPaths {
			if _, err = os.Stat(manifestPath); err == nil {
				return manifestPath, err
			}
		}
		return "", err
	} else {
		return userSpecifiedPath, nil
	}
}
