package configv3

import (
	"sort"
	"strings"
)

const (
	// DefaultPluginRepoName is the name of the preinstalled plugin repository.
	DefaultPluginRepoName = "CF-Community"

	// DefaultPluginRepoURL is the URL of the preinstalled plugin repository.
	DefaultPluginRepoURL = "https://plugins.cloudfoundry.org"
)

// PluginRepository is a saved plugin repository
type PluginRepository struct {
	Name string `json:"Name"`
	URL  string `json:"URL"`
}

// PluginRepositories returns the currently configured plugin repositories from the
// .cf/config.json
func (config *Config) PluginRepositories() []PluginRepository {
	repos := config.ConfigFile.PluginRepositories
	sort.Slice(repos, func(i, j int) bool {
		return strings.ToLower(repos[i].Name) < strings.ToLower(repos[j].Name)
	})
	return repos
}
