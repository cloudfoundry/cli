package manifestparser

import (
	"errors"
	"fmt"
	"os"
	"path"
)

func Validate(manifestParser *Parser) error {
	var err error

	initialDir, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(path.Dir(manifestParser.PathToManifest))
	if err != nil {
		return err
	}

	for _, application := range manifestParser.Applications {
		if application.Path != "" {
			_, err = os.Stat(application.Path)
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

	err = os.Chdir(initialDir)
	if err != nil {
		return err
	}

	return nil
}

