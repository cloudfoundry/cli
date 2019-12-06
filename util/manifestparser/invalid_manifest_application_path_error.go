package manifestparser

import "fmt"

type InvalidManifestApplicationPathError struct {
	Path string
}

func (e InvalidManifestApplicationPathError) Error() string {
	return fmt.Sprintf("File not found locally, make sure the file exists at given path %s", e.Path)
}
