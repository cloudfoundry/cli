package configv3

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/util"
)

// PluginsConfig represents the plugin configuration
type PluginsConfig struct {
	Plugins map[string]Plugin `json:"Plugins"`
}

// Plugin represents the plugin as a whole, not be confused with PluginCommand
type Plugin struct {
	Name     string
	Location string          `json:"Location"`
	Version  PluginVersion   `json:"Version"`
	Commands []PluginCommand `json:"Commands"`
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

// PluginCommand represents an individual command inside a plugin
type PluginCommand struct {
	Name         string             `json:"Name"`
	Alias        string             `json:"Alias"`
	HelpText     string             `json:"HelpText"`
	UsageDetails PluginUsageDetails `json:"UsageDetails"`
}

// PluginUsageDetails contains the usage metadata provided by the plugin
type PluginUsageDetails struct {
	Usage   string            `json:"Usage"`
	Options map[string]string `json:"Options"`
}

// Plugins returns installed plugins from the config sorted by name (case-insensitive).
func (config *Config) Plugins() []Plugin {
	plugins := []Plugin{}
	for _, plugin := range config.pluginsConfig.Plugins {
		plugins = append(plugins, plugin)
	}
	sort.Slice(plugins, func(i, j int) bool {
		return strings.ToLower(plugins[i].Name) < strings.ToLower(plugins[j].Name)
	})
	return plugins
}

// CalculateSHA1 returns the sha1 value of the plugin executable. If an error
// is encountered calculating sha1, N/A is returned
func (p Plugin) CalculateSHA1() string {
	fileSHA, err := util.NewSha1Checksum(p.Location).ComputeFileSha1()

	if err != nil {
		return "N/A"
	}

	return fmt.Sprintf("%x", fileSHA)
}

// PluginCommands returns the plugin's commands sorted by command name.
func (p Plugin) PluginCommands() []PluginCommand {
	sort.Slice(p.Commands, func(i, j int) bool {
		return strings.ToLower(p.Commands[i].Name) < strings.ToLower(p.Commands[j].Name)
	})
	return p.Commands
}

// CommandName returns the name of the plugin. The name is concatenated with
// alias if alias is specified.
func (c PluginCommand) CommandName() string {
	if c.Name != "" && c.Alias != "" {
		return fmt.Sprintf("%s, %s", c.Name, c.Alias)
	}
	return c.Name
}

// PluginHome returns the plugin configuration directory to:
//   1. The $CF_PLUGIN_HOME/.cf/plugins environment variable if set
//   2. Defaults to the home directory (outlined in LoadConfig)/.cf/plugins
func (config *Config) PluginHome() string {
	if config.ENV.CFPluginHome != "" {
		return filepath.Join(config.ENV.CFPluginHome, ".cf", "plugins")
	}

	return filepath.Join(homeDirectory(), ".cf", "plugins")
}

// AddPlugin adds the specified plugin to PluginsConfig
func (config *Config) AddPlugin(plugin Plugin) {
	config.pluginsConfig.Plugins[plugin.Name] = plugin
}

// RemovePlugin removes the specified plugin from PluginsConfig idempotently
func (config *Config) RemovePlugin(pluginName string) {
	delete(config.pluginsConfig.Plugins, pluginName)
}

// GetPlugin returns the requested plugin and true if it exists.
func (config *Config) GetPlugin(pluginName string) (Plugin, bool) {
	plugin, exists := config.pluginsConfig.Plugins[pluginName]
	return plugin, exists
}

// GetPluginCaseInsensitive finds the first matching plugin name case
// insensitive and returns true if it exists.
func (config *Config) GetPluginCaseInsensitive(pluginName string) (Plugin, bool) {
	for name, plugin := range config.pluginsConfig.Plugins {
		if strings.ToLower(name) == strings.ToLower(pluginName) {
			return plugin, true
		}
	}

	return Plugin{}, false
}

// WritePluginConfig writes the plugin config to config.json in the plugin home
// directory.
func (config *Config) WritePluginConfig() error {
	// Marshal JSON
	rawConfig, err := json.MarshalIndent(config.pluginsConfig, "", "  ")
	if err != nil {
		return err
	}

	pluginFileDir := filepath.Join(config.PluginHome())
	err = os.MkdirAll(pluginFileDir, 0700)
	if err != nil {
		return err
	}

	// Write to file
	return ioutil.WriteFile(filepath.Join(pluginFileDir, "config.json"), rawConfig, 0600)
}
