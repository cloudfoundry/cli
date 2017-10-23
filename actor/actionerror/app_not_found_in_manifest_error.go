package actionerror

import "fmt"

type AppNotFoundInManifestError struct {
	Name string
}

func (e AppNotFoundInManifestError) Error() string {
	return fmt.Sprintf("specfied app: %s not found in manifest", e.Name)
}
