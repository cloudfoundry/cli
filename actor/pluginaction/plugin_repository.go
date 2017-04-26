package pluginaction

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/api/plugin"
)

type PluginRepository plugin.PluginRepository

type RepositoryNameTakenError struct {
	Name string
}

func (e RepositoryNameTakenError) Error() string {
	return fmt.Sprintf("Plugin repo named '%s' already exists, please use another name.", e.Name)
}

type RepositoryURLTakenError struct {
	Name string
	URL  string
}

func (e RepositoryURLTakenError) Error() string {
	return fmt.Sprintf("%s (%s) already exists.", e.URL, e.Name)
}

type AddPluginRepositoryError struct {
	Name    string
	URL     string
	Message string
}

func (e AddPluginRepositoryError) Error() string {
	return fmt.Sprintf("Could not add repository '%s' from %s: %s", e.Name, e.URL, e.Message)
}

func (actor Actor) AddPluginRepository(repoName string, repoURL string) error {
	normalizedURL, err := normalizeURLPath(repoURL)
	if err != nil {
		return AddPluginRepositoryError{
			Name:    repoName,
			URL:     repoURL,
			Message: err.Error(),
		}
	}

	for _, repository := range actor.config.PluginRepositories() {
		if repoName == repository.Name {
			return RepositoryNameTakenError{Name: repository.Name}
		} else if normalizedURL == repository.URL {
			return RepositoryURLTakenError{
				Name: repository.Name,
				URL:  repository.URL,
			}
		}
	}

	_, err = actor.client.GetPluginRepository(normalizedURL)
	if err != nil {
		return AddPluginRepositoryError{
			Name:    repoName,
			URL:     normalizedURL,
			Message: err.Error(),
		}
	}

	actor.config.AddPluginRepository(repoName, normalizedURL)

	return nil
}

func normalizeURLPath(rawURL string) (string, error) {
	prefix := ""
	if !strings.Contains(rawURL, "://") {
		prefix = "https://"
	}

	normalizedURL := fmt.Sprintf("%s%s", prefix, rawURL)

	return strings.TrimSuffix(normalizedURL, "/"), nil
}
