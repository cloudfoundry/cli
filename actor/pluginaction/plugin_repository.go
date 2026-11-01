package pluginaction

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/api/plugin"
	"code.cloudfoundry.org/cli/v8/util/configv3"
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

// RemovePluginRepository removes the registered repository whose name matches
// repositoryName (trimmed, case-insensitive).
func (actor Actor) RemovePluginRepository(repositoryName string) error {
	repository, err := actor.GetPluginRepository(strings.TrimSpace(repositoryName))
	if err != nil {
		return err
	}

	actor.config.RemovePluginRepository(repository.Name)
	return nil
}

// RepositoryPlugins lists the plugins available in a plugin repository.
type RepositoryPlugins struct {
	RepositoryName string
	Plugins        []plugin.Plugin
}

// GetPluginsFromRepositories fetches the plugins of each repository, in order.
// Repositories that cannot be fetched are skipped and reported as errors.
func (actor Actor) GetPluginsFromRepositories(repositories []configv3.PluginRepository) ([]RepositoryPlugins, []error) {
	var (
		repositoryPlugins []RepositoryPlugins
		errs              []error
	)

	for _, repository := range repositories {
		repositoryURL := repository.URL
		if repositoryURL == "http://plugins.cloudfoundry.org" {
			repositoryURL = configv3.DefaultPluginRepoURL
		}

		pluginRepository, err := actor.client.GetPluginRepository(repositoryURL)
		if err != nil {
			errs = append(errs, actionerror.FetchingPluginInfoFromRepositoryError{
				RepositoryName: repository.Name,
				Err:            err,
			})
			continue
		}

		repositoryPlugins = append(repositoryPlugins, RepositoryPlugins{
			RepositoryName: repository.Name,
			Plugins:        pluginRepository.Plugins,
		})
	}

	return repositoryPlugins, errs
}

func normalizeURLPath(rawURL string) (string, error) {
	prefix := ""
	if !strings.Contains(rawURL, "://") {
		prefix = "https://"
	}

	normalizedURL := fmt.Sprintf("%s%s", prefix, rawURL)

	return strings.TrimSuffix(normalizedURL, "/"), nil
}
