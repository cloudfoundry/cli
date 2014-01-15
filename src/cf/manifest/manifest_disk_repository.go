package manifest

import (
	"errors"
	"generic"
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

	if os.Getenv("CF_MANIFEST") != "true" {
		return
	}

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
	mergedMap = generic.NewMap()

	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		dir, _ := os.Getwd()
		println("expected file to exist, but it does not: ", path, dir)
		return
	}
	defer file.Close()

	mapp, err := Parse(file)
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

	mergedMap = generic.Reduce([]generic.Map{inheritedMap, mapp}, mergedMap, reducer)
	return
}

func reducer(key, val interface{}, reduced generic.Map) generic.Map {
	switch {
	case reduced.Has(key) == false:
		reduced.Set(key, val)
		return reduced

	case generic.IsMappable(val):
		maps := []generic.Map{generic.NewMap(reduced.Get(key)), generic.NewMap(val)}
		mergedMap := generic.Reduce(maps, generic.NewMap(), reducer)
		reduced.Set(key, mergedMap)
		return reduced

	case generic.IsSliceable(val):
		reduced.Set(key, append(reduced.Get(key).([]interface{}), val.([]interface{})...))
		return reduced

	default:
		reduced.Set(key, val)
		return reduced
	}
}

func (repo ManifestDiskRepository) ManifestPath(userSpecifiedPath string) (manifestDir, manifestFilename string, err error) {
	if userSpecifiedPath == "" {
		userSpecifiedPath, err = os.Getwd()
		if err != nil {
			err = errors.New("Error finding current directory: "+err.Error())
			return
		}
	}

	fileInfo, err := os.Stat(userSpecifiedPath)
	if err != nil {
		err = errors.New("Error finding manifest path: "+err.Error())
		return
	}

	if fileInfo.IsDir() {
		manifestDir = userSpecifiedPath
		manifestFilename = "manifest.yml"
	} else {
		manifestDir = filepath.Dir(userSpecifiedPath)
		manifestFilename = fileInfo.Name()
	}

	fileInfo, err = os.Stat(userSpecifiedPath)
	return
}
