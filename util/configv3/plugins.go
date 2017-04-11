package configv3

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"code.cloudfoundry.org/cli/util/sorting"
)

const (
	// DefaultPluginRepoName is the name of the preinstalled plugin repository.
	DefaultPluginRepoName = "CF-Community"

	// DefaultPluginRepoURL is the URL of the preinstalled plugin repository.
	DefaultPluginRepoURL = "https://plugins.cloudfoundry.org"
)

// PluginRepos is a saved plugin repository
type PluginRepos struct {
	Name string `json:"Name"`
	URL  string `json:"URL"`
}

// PluginsConfig represents the plugin configuration
type PluginsConfig struct {
	Plugins map[string]Plugin `json:"Plugins"`
}

// Plugin represents the plugin as a whole, not be confused with PluginCommand
type Plugin struct {
	Location string         `json:"Location"`
	Version  PluginVersion  `json:"Version"`
	Commands PluginCommands `json:"Commands"`
}

// CalculateSHA1 returns the sha1 value of the plugin executable. If an error
// is encountered calculating sha1, N/A is returned
func (p Plugin) CalculateSHA1() string {
	fileSHA := ""
	contents, err := ioutil.ReadFile(p.Location)
	if err != nil {
		fileSHA = "N/A"
	} else {
		fileSHA = fmt.Sprintf("%x", sha1.New().Sum(contents))
	}
	return fileSHA
}

// PluginVersion is the plugin version information
type PluginVersion struct {
	Major int `json:"Major"`
	Minor int `json:"Minor"`
	Build int `json:"Build"`
}

// String returns the plugin's version in the format x.y.z.
func (v PluginVersion) String() string {
	if v.Major == 0 && v.Minor == 0 && v.Build == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Build)
}

// PluginCommands is a list of plugins that implements the sort.Interface
type PluginCommands []PluginCommand

func (p PluginCommands) Len() int               { return len(p) }
func (p PluginCommands) Swap(i int, j int)      { p[i], p[j] = p[j], p[i] }
func (p PluginCommands) Less(i int, j int) bool { return sorting.SortAlphabetic(p[i].Name, p[j].Name) }

// PluginCommand represents an individual command inside a plugin
type PluginCommand struct {
	Name         string             `json:"Name"`
	Alias        string             `json:"Alias"`
	HelpText     string             `json:"HelpText"`
	UsageDetails PluginUsageDetails `json:"UsageDetails"`
}

// CommandName returns the name of the plugin. The name is concatenated with
// alias if alias is specified.
func (c PluginCommand) CommandName() string {
	if c.Name != "" && c.Alias != "" {
		return fmt.Sprintf("%s, %s", c.Name, c.Alias)
	}
	return c.Name
}

// PluginUsageDetails contains the usage metadata provided by the plugin
type PluginUsageDetails struct {
	Usage   string            `json:"Usage"`
	Options map[string]string `json:"Options"`
}

// PluginHome returns the plugin configuration directory based off:
//   1. The $CF_PLUGIN_HOME environment variable if set
//   2. Defaults to the home diretory (outlined in LoadConfig)/.cf/plugins
func (config *Config) PluginHome() string {
	if config.ENV.CFPluginHome != "" {
		return filepath.Join(config.ENV.CFPluginHome, ".cf", "plugins")
	}

	return filepath.Join(homeDirectory(), ".cf", "plugins")
}

// Plugins returns back the plugin configuration read from the plugin home
func (config *Config) Plugins() map[string]Plugin {
	return config.pluginConfig.Plugins
}

// PluginRepos returns the currently configured plugin repositories from the
// .cf/config.json
func (config *Config) PluginRepos() []PluginRepos {
	return config.ConfigFile.PluginRepos
}
