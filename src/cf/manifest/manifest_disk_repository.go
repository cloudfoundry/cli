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
	ReadManifest(path string) (manifest *Manifest, errs ManifestErrors)
	ManifestPath(userSpecifiedPath string) (manifestDir, manifestFilename string, err error)
}

type ManifestDiskRepository struct {
}

func NewManifestDiskRepository() (repo ManifestRepository) {
	return ManifestDiskRepository{}
}

func (repo ManifestDiskRepository) ReadManifest(path string) (m *Manifest, errs ManifestErrors) {
	m = NewEmptyManifest()

	mapp, err := repo.readAllYAMLFiles(path)
	if err != nil {
		errs = append(errs, err)
		return
	}

	m, errs = NewManifest(mapp)
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

	yamlMap = generic.NewMap(document)
	return
}

func (repo ManifestDiskRepository) ManifestPath(userSpecifiedPath string) (manifestDir, manifestFilename string, err error) {
	if userSpecifiedPath == "" {
		userSpecifiedPath, err = os.Getwd()
		if err != nil {
			err = errors.New("Error finding current directory: " + err.Error())
			return
		}
	}

	fileInfo, err := os.Stat(userSpecifiedPath)
	if err != nil {
		err = errors.New("Error finding manifest path: " + err.Error())
		return
	}

	if fileInfo.IsDir() {
		manifestDir = userSpecifiedPath
		manifestFilename = "manifest.yml"
	} else {
		manifestDir = filepath.Dir(userSpecifiedPath)
		manifestFilename = fileInfo.Name()
	}

	return
}
