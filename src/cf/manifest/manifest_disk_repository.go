package manifest

import (
	"errors"
	"generic"
	"github.com/cloudfoundry/gamble"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ManifestRepository interface {
	ReadManifest(string) (manifest *Manifest, path string, errors ManifestErrors)
}

type ManifestDiskRepository struct{}

func NewManifestDiskRepository() (repo ManifestRepository) {
	return ManifestDiskRepository{}
}

func (repo ManifestDiskRepository) ReadManifest(inputPath string) (m *Manifest, manifestPath string, errs ManifestErrors) {
	m = NewEmptyManifest()

	basePath, fileName, err := repo.manifestPath(inputPath)
	if err != nil {
		errs = append(errs, errors.New("Error finding manifest: "+err.Error()))
		return
	}

	manifestPath = filepath.Join(basePath, fileName)

	mapp, err := repo.readAllYAMLFiles(manifestPath)
	if err != nil {
		errs = append(errs, err)
		return
	}

	m, errs = NewManifest(basePath, mapp)
	if !errs.Empty() {
		return
	}

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
	yamlBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	document, err := gamble.Parse(string(yamlBytes))
	if err != nil {
		return
	}

	if !generic.IsMappable(document) {
		err = errors.New("Invalid manifest. Expected a map")
		return
	}

	yamlMap = generic.NewMap(document)
	return
}

func (repo ManifestDiskRepository) manifestPath(userSpecifiedPath string) (manifestDir, manifestFilename string, err error) {
	fileInfo, err := os.Stat(userSpecifiedPath)
	if err != nil {
		return
	}

	if fileInfo.IsDir() {
		manifestDir = userSpecifiedPath
		manifestFilename = "manifest.yml"
		_, err = os.Stat(filepath.Join(manifestDir, manifestFilename))
		if err != nil {
			return
		}
	} else {
		manifestDir = filepath.Dir(userSpecifiedPath)
		manifestFilename = fileInfo.Name()
	}

	return
}
