package pluginaction

import (
	"fmt"
	"strings"
)

type RepositoryAlreadyExistsError struct {
	Name string
	URL  string
}

func (e RepositoryAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already registered as %s.", e.URL, e.Name)
}

type RepositoryNameTakenError struct {
	Name string
}

func (e RepositoryNameTakenError) Error() string {
	return fmt.Sprintf("Plugin repo named '%s' already exists, please use another name.", e.Name)
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

	repoNameLowerCased := strings.ToLower(repoName)
	for _, repository := range actor.config.PluginRepositories() {
		existingRepoNameLowerCased := strings.ToLower(repository.Name)
		switch {
		case repoNameLowerCased == existingRepoNameLowerCased && normalizedURL == repository.URL:
			return RepositoryAlreadyExistsError{Name: repository.Name, URL: repository.URL}
		case repoNameLowerCased == existingRepoNameLowerCased && normalizedURL != repository.URL:
			return RepositoryNameTakenError{Name: repository.Name}
		case repoNameLowerCased != existingRepoNameLowerCased:
			continue
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
