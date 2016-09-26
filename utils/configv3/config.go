// Package config package contains everything related to the CF CLI Configuration.
package configv3

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	DefaultStagingTimeout = 15 * time.Minute
	DefaultStartupTimeout = 5 * time.Minute
	// DefaultPingerThrottle = 5 * time.Second

	DefaultTarget = ""
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
