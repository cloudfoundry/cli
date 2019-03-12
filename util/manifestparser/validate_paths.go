package manifestparser

import (
	"os"
	"path/filepath"
)

type InvalidManifestApplicationPathError struct {
	Path string
}

func (InvalidManifestApplicationPathError) Error() string {
	return "Path in manifest is invalid"
}

func ValidatePaths(manifestParser Parser) error {
	var err error

	for _, application := range manifestParser.Applications {
		if application.Path != "" {

			if filepath.IsAbs(application.Path) {
				_, err = os.Stat(application.Path)
			} else {
				manifestPath := filepath.Dir(manifestParser.PathToManifest)
				_, err = os.Stat(filepath.Join(manifestPath, application.Path))
			}

			if err != nil {
				if os.IsNotExist(err) {
					return InvalidManifestApplicationPathError{
						Path: application.Path,
					}
				}
			}
		}
	}

	return nil
}
