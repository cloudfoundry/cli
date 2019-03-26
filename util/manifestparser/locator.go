package manifestparser

import (
	"os"
	"path/filepath"
)

type Locator struct {
	FilesToCheckFor []string
}

func NewLocator() *Locator {
	return &Locator{
		FilesToCheckFor: []string{
			"manifest.yml",
			"manifest.yaml",
		},
	}
}

func (loc Locator) Path(filepathOrDirectory string) (string, bool, error) {
	info, err := os.Stat(filepathOrDirectory)
	if os.IsNotExist(err) {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	}

	if info.IsDir() {
		return loc.handleDir(filepathOrDirectory)
	}

	return loc.handleFilepath(filepathOrDirectory)
}

func (loc Locator) handleDir(dir string) (string, bool, error) {
	for _, filename := range loc.FilesToCheckFor {
		fullPath := filepath.Join(dir, filename)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, true, nil
		} else if !os.IsNotExist(err) {
			return "", false, err
		}
	}

	return "", false, nil
}

func (Locator) handleFilepath(filepath string) (string, bool, error) {
	return filepath, true, nil
}
