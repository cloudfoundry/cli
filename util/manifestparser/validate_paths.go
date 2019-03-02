package manifestparser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

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
					return errors.New(
						fmt.Sprintf(
							"Path '%s' does not exist for application '%s' in manifest",
							application.Path,
							application.Name,
						),
					)
				}
			}
		}
	}

	return nil
}
