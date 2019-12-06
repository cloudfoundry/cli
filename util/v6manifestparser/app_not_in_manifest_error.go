package v6manifestparser

import "fmt"

type AppNotInManifestError struct {
	Name string
}

func (e AppNotInManifestError) Error() string {
	return fmt.Sprintf("Could not find app named '%s' in manifest", e.Name)
}
