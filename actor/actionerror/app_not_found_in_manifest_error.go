package actionerror

import "fmt"

type AppNotFoundInManifestError struct {
	Name string
}

func (e AppNotFoundInManifestError) Error() string {
	return fmt.Sprintf("specified app: %s not found in manifest", e.Name)
}
