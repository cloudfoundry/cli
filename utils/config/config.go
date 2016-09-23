// Package config package contains everything related to the CF CLI Configuration.
package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/utils/sortutils"
)

const (
	DefaultStagingTimeout = 15 * time.Minute
	DefaultStartupTimeout = 5 * time.Minute
	// DefaultPingerThrottle = 5 * time.Second

	DefaultTarget         = ""
	DefaultColorEnabled   = "true"
	DefaultLocale         = ""
	DefaultPluginRepoName = "CF-Community"
	DefaultPluginRepoURL  = "https://plugins.cloudfoundry.org"
)

// LoadConfig loads the config from the .cf/config.json and os.ENV. If the
// config.json does not exists, it will use a default config in it's place.
//
// The '.cf' directory will be read in one of the following locations on UNIX
// Systems:
//   1. $CF_HOME/.cf if $CF_HOME is set
//   2. $HOME/.cf as the default
//
// The '.cf' directory will be read in one of the following locations on
// Windows Systems:
//   1. CF_HOME\.cf if CF_HOME is set
//   2. HOMEDRIVE\HOMEPATH\.cf if HOMEDRIVE or HOMEPATH is set
//   3. USERPROFILE\.cf as the default
func LoadConfig() (*Config, error) {
	filePath := ConfigFilePath()

	var config Config
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		config = Config{
			ConfigFile: CFConfig{
				ConfigVersion: 3,
				Target:        DefaultTarget,
				ColorEnabled:  DefaultColorEnabled,
				PluginRepos: []PluginRepos{{
					Name: DefaultPluginRepoName,
					URL:  DefaultPluginRepoURL,
				}},
			},
		}
	} else {
		file, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(file, &config.ConfigFile)
		if err != nil {
			return nil, err
		}
	}

	config.ENV = EnvOverride{
		BinaryName:       os.Args[0],
		CFColor:          os.Getenv("CF_COLOR"),
		CFPluginHome:     os.Getenv("CF_PLUGIN_HOME"),
		CFStagingTimeout: os.Getenv("CF_STAGING_TIMEOUT"),
		CFStartupTimeout: os.Getenv("CF_STARTUP_TIMEOUT"),
		CFTrace:          os.Getenv("CF_TRACE"),
		HTTPSProxy:       os.Getenv("https_proxy"),
		Lang:             os.Getenv("LANG"),
		LCAll:            os.Getenv("LC_ALL"),
		Experimental:     os.Getenv("EXPERIMENTAL"),
	}

	pluginFilePath := filepath.Join(config.PluginHome(), "config.json")
	if _, err := os.Stat(pluginFilePath); os.IsNotExist(err) {
		config.pluginConfig = PluginsConfig{}
	} else {
		file, err := ioutil.ReadFile(pluginFilePath)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(file, &config.pluginConfig)
		if err != nil {
			return nil, err
		}
	}

	return &config, nil
}

// WriteConfig creates the .cf directory and then writes the config.json. The
// location of .cf directory is written in the same way LoadConfig reads .cf
// directory.
func WriteConfig(c *Config) error {
	rawConfig, err := json.MarshalIndent(c.ConfigFile, "", "  ")
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(homeDirectory(), ".cf"), 0700)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(ConfigFilePath(), rawConfig, 0600)
}

// Config combines the settings taken from the .cf/config.json, os.ENV, and the
// plugin config.
type Config struct {
	// ConfigFile stores the configuration from the .cf/config
	ConfigFile CFConfig

	// ENV stores the configuration from os.ENV
	ENV EnvOverride

	pluginConfig PluginsConfig
}

// CFConfig represents .cf/config.json
type CFConfig struct {
	ConfigVersion            int           `json:"ConfigVersion"`
	Target                   string        `json:"Target"`
	APIVersion               string        `json:"APIVersion"`
	AuthorizationEndpoint    string        `json:"AuthorizationEndpoint"`
	LoggregatorEndpoint      string        `json:"LoggregatorEndPoint"`
	DopplerEndpoint          string        `json:"DopplerEndPoint"`
	UAAEndpoint              string        `json:"UaaEndpoint"`
	RoutingEndpoint          string        `json:"RoutingAPIEndpoint"`
	AccessToken              string        `json:"AccessToken"`
	SSHOAuthClient           string        `json:"SSHOAuthClient"`
	RefreshToken             string        `json:"RefreshToken"`
	TargetedOrganization     Organization  `json:"OrganizationFields"`
	TargetedSpace            Space         `json:"SpaceFields"`
	SkipSSLValidation        bool          `json:"SSLDisabled"`
	AsyncTimeout             int           `json:"AsyncTimeout"`
	Trace                    string        `json:"Trace"`
	ColorEnabled             string        `json:"ColorEnabled"`
	Locale                   string        `json:"Locale"`
	PluginRepos              []PluginRepos `json:"PluginRepos"`
	MinCLIVersion            string        `json:"MinCLIVersion"`
	MinRecommendedCLIVersion string        `json:"MinRecommendedCLIVersion"`
}

// Organization contains basic information about the targeted organization
type Organization struct {
	GUID            string          `json:"GUID"`
	Name            string          `json:"Name"`
	QuotaDefinition QuotaDefinition `json:"QuotaDefinition"`
}

// QuotaDefinition contains information about the organization's quota
type QuotaDefinition struct {
	GUID                    string `json:"guid"`
	Name                    string `json:"name"`
	MemoryLimit             int    `json:"memory_limit"`
	InstanceMemoryLimit     int    `json:"instance_memory_limit"`
	TotalRoutes             int    `json:"total_routes"`
	TotalServices           int    `json:"total_services"`
	NonBasicServicesAllowed bool   `json:"non_basic_services_allowed"`
	AppInstanceLimit        int    `json:"app_instance_limit"`
	TotalReservedRoutePorts int    `json:"total_reserved_route_ports"`
}

// Space contains basic information about the targeted space
type Space struct {
	GUID     string `json:"GUID"`
	Name     string `json:"Name"`
	AllowSSH bool   `json:"AllowSSH"`
}

// PluginRepos is a saved plugin repository
type PluginRepos struct {
	Name string `json:"Name"`
	URL  string `json:"URL"`
}

// EnvOverride represents all the environment variables read by the CF CLI
type EnvOverride struct {
	BinaryName       string
	CFColor          string
	CFHome           string
	CFPluginHome     string
	CFStagingTimeout string
	CFStartupTimeout string
	CFTrace          string
	HTTPSProxy       string
	Lang             string
	LCAll            string
	Experimental     string
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

// PluginVersion is the plugin version information
type PluginVersion struct {
	Major int `json:"Major"`
	Minor int `json:"Minor"`
	Build int `json:"Build"`
}

// PluginCommands is a list of plugins that implements the sort.Interface
type PluginCommands []PluginCommand

func (p PluginCommands) Len() int               { return len(p) }
func (p PluginCommands) Swap(i int, j int)      { p[i], p[j] = p[j], p[i] }
func (p PluginCommands) Less(i int, j int) bool { return sortutils.SortAlphabetic(p[i].Name, p[j].Name) }

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

// ColorSetting is a trinary operator that represents if the display should
// have colors enabled, disabled, or automatically detected.
type ColorSetting int

const (
	// ColorDisbled means that no colors/bolding will be displayed
	ColorDisbled ColorSetting = iota
	// ColorEnabled means colors/bolding will be displayed
	ColorEnabled
	// ColorAuto means that the UI should decide if colors/bolding will be
	// enabled
	ColorAuto
)

// ColorEnabled returns the color setting based off:
//   1. The $CF_COLOR environment variable if set (0/1/t/f/true/false)
//   2. The 'ColorEnabled' value in the .cf/config.json if set
//   3. Defaults to ColorEnabled if nothing is set
func (config *Config) ColorEnabled() ColorSetting {
	if config.ENV.CFColor != "" {
		val, err := strconv.ParseBool(config.ENV.CFColor)
		if err == nil {
			return config.boolToColorSetting(val)
		}
	}

	val, err := strconv.ParseBool(config.ConfigFile.ColorEnabled)
	if err != nil {
		return ColorEnabled
	}
	return config.boolToColorSetting(val)
}

func (config *Config) boolToColorSetting(val bool) ColorSetting {
	if val {
		return ColorEnabled
	}

	return ColorDisbled
}

// Target returns the CC API URL
func (config *Config) Target() string {
	return config.ConfigFile.Target
}

// APIVersion returns the CC API Version
func (config *Config) APIVersion() string {
	return config.ConfigFile.APIVersion
}

// TargetedOrganization returns the currently targeted organization
func (config *Config) TargetedOrganization() Organization {
	return config.ConfigFile.TargetedOrganization
}

// TargetedSpace returns the currently targeted space
func (config *Config) TargetedSpace() Space {
	return config.ConfigFile.TargetedSpace
}

// CurrentUser returns user information decoded from the JWT access token in
// .cf/config.json
func (config *Config) CurrentUser() (User, error) {
	return decodeUserFromJWT(config.ConfigFile.AccessToken)
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

// StagingTimeout returns the max time an application staging should take. The
// time is based off of:
//   1. The $CF_STAGING_TIMEOUT environment variable if set
//   2. Defaults to the DefaultStagingTimeout
func (config *Config) StagingTimeout() time.Duration {
	if config.ENV.CFStagingTimeout != "" {
		val, err := strconv.ParseInt(config.ENV.CFStagingTimeout, 10, 64)
		if err == nil {
			return time.Duration(val) * time.Minute
		}
	}

	return DefaultStagingTimeout
}

// StartupTimeout returns the max time an application should take to start. The
// time is based off of:
//   1. The $CF_STARTUP_TIMEOUT environment variable if set
//   2. Defaults to the DefaultStartupTimeout
func (config *Config) StartupTimeout() time.Duration {
	if config.ENV.CFStartupTimeout != "" {
		val, err := strconv.ParseInt(config.ENV.CFStartupTimeout, 10, 64)
		if err == nil {
			return time.Duration(val) * time.Minute
		}
	}

	return DefaultStartupTimeout
}

// HTTPSProxy returns the proxy url that the CLI should use. The url is based
// off of:
//   1. The $https_proxy environment variable if set
//   2. Defaults to the empty string
func (config *Config) HTTPSProxy() string {
	if config.ENV.HTTPSProxy != "" {
		return config.ENV.HTTPSProxy
	}

	return ""
}

// Locale returns the locale/language the UI should be displayed in. This value
// is based off of:
//   1. The 'Locale' setting in the .cf/config.json
//   2. The $LC_ALL environment variable if set
//   3. The $LANG environment variable if set
//   4. Defaults to DefaultLocale
func (config *Config) Locale() string {
	if config.ConfigFile.Locale != "" {
		return config.ConfigFile.Locale
	}

	if config.ENV.LCAll != "" {
		return config.convertLocale(config.ENV.LCAll)
	}

	if config.ENV.Lang != "" {
		return config.convertLocale(config.ENV.Lang)
	}

	return DefaultLocale
}

// BinaryName returns the running name of the CF CLI
func (config *Config) BinaryName() string {
	return config.ENV.BinaryName
}

// Experimental returns whether or not to run experimental CLI commands
func (config *Config) Experimental() bool {
	envValStr := config.ENV.Experimental

	if envValStr != "" {
		envVal, err := strconv.ParseBool(envValStr)
		if err == nil {
			return envVal
		}
	}

	return false
}

// SetOrganizationInformation sets the currently targeted organization
func (config *Config) SetOrganizationInformation(guid string, name string) {
	config.ConfigFile.TargetedOrganization.GUID = guid
	config.ConfigFile.TargetedOrganization.Name = name
	config.ConfigFile.TargetedOrganization.QuotaDefinition = QuotaDefinition{}
}

// SetSpaceInformation sets the currently targeted space
func (config *Config) SetSpaceInformation(guid string, name string, allowSSH bool) {
	config.ConfigFile.TargetedSpace.GUID = guid
	config.ConfigFile.TargetedSpace.Name = name
	config.ConfigFile.TargetedSpace.AllowSSH = allowSSH
}

// SetTargetInformation sets the currently targeted CC API and related other
// related API URLs
func (config *Config) SetTargetInformation(api string, apiVersion string, auth string, loggregator string, doppler string, uaa string, routing string, skipSSLValidation bool) {
	config.ConfigFile.Target = api
	config.ConfigFile.APIVersion = apiVersion
	config.ConfigFile.AuthorizationEndpoint = auth
	config.ConfigFile.LoggregatorEndpoint = loggregator
	config.ConfigFile.DopplerEndpoint = doppler
	config.ConfigFile.UAAEndpoint = uaa
	config.ConfigFile.RoutingEndpoint = routing
	config.ConfigFile.SkipSSLValidation = skipSSLValidation

	config.SetOrganizationInformation("", "")
	config.SetSpaceInformation("", "", false)
}

// SetTokenInformation sets the current token/user information
func (config *Config) SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string) {
	config.ConfigFile.AccessToken = accessToken
	config.ConfigFile.RefreshToken = refreshToken
	config.ConfigFile.SSHOAuthClient = sshOAuthClient
}

// PluginRepos returns the currently configured plugin repositories from the
// .cf/config.json
func (config *Config) PluginRepos() []PluginRepos {
	return config.ConfigFile.PluginRepos
}

func (config *Config) convertLocale(local string) string {
	lang := strings.Split(local, ".")[0]
	return strings.Replace(lang, "_", "-", -1)
}
