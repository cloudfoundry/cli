package actionerror

import "fmt"

// FetchingPluginInfoFromRepositoryError is returned an error is encountered
// getting plugin info from a repository.
type FetchingPluginInfoFromRepositoryError struct {
	RepositoryName string
	Err            error
}

func (e FetchingPluginInfoFromRepositoryError) Error() string {
	return fmt.Sprintf("Plugin repository %s returned %s.", e.RepositoryName, e.Err.Error())
}
