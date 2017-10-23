package pluginaction

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/util/configv3"
)

func (actor Actor) AddPluginRepository(repoName string, repoURL string) error {
	normalizedURL, err := normalizeURLPath(repoURL)
	if err != nil {
		return actionerror.AddPluginRepositoryError{
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
			return actionerror.RepositoryAlreadyExistsError{Name: repository.Name, URL: repository.URL}
		case repoNameLowerCased == existingRepoNameLowerCased && normalizedURL != repository.URL:
			return actionerror.RepositoryNameTakenError{Name: repository.Name}
		case repoNameLowerCased != existingRepoNameLowerCased:
			continue
		}
	}

	_, err = actor.client.GetPluginRepository(normalizedURL)
	if err != nil {
		return actionerror.AddPluginRepositoryError{
			Name:    repoName,
			URL:     normalizedURL,
			Message: err.Error(),
		}
	}

	actor.config.AddPluginRepository(repoName, normalizedURL)
	return nil
}

func (actor Actor) GetPluginRepository(repositoryName string) (configv3.PluginRepository, error) {
	repositoryNameLowered := strings.ToLower(repositoryName)

	for _, repository := range actor.config.PluginRepositories() {
		if repositoryNameLowered == strings.ToLower(repository.Name) {
			return repository, nil
		}
	}
	return configv3.PluginRepository{}, actionerror.RepositoryNotRegisteredError{Name: repositoryName}
}

func (actor Actor) IsPluginRepositoryRegistered(repositoryName string) bool {
	for _, repository := range actor.config.PluginRepositories() {
		if repositoryName == repository.Name {
			return true
		}
	}
	return false
}

func normalizeURLPath(rawURL string) (string, error) {
	prefix := ""
	if !strings.Contains(rawURL, "://") {
		prefix = "https://"
	}

	normalizedURL := fmt.Sprintf("%s%s", prefix, rawURL)

	return strings.TrimSuffix(normalizedURL, "/"), nil
}
