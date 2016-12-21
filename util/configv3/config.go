// Package configv3 package contains everything related to the CF CLI Configuration.
package configv3

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"code.cloudfoundry.org/cli/version"
)

const (
	// DefaultStagingTimeout is the default timeout for application staging.
	DefaultStagingTimeout = 15 * time.Minute

	// DefaultStartupTimeout is the default timeout for application starting.
	DefaultStartupTimeout = 5 * time.Minute
	// DefaultPingerThrottle = 5 * time.Second

	// DefaultDialTimeout is the default timeout for the dail.
	DefaultDialTimeout = 5 * time.Second

	// DefaultTarget is the default CFConfig value for Target.
	DefaultTarget = ""

	// DefaultUAAOAuthClient is the default client ID for the CLI when
	// communicating with the UAA.
	DefaultUAAOAuthClient = "cf"

	// DefaultCFOClientSecret is the default client secret for the CLI when
	// communicating with the UAA.
	DefaultUAAOAuthClientSecret = ""
)

// LoadConfig loads the config from the .cf/config.json and os.ENV. If the
// config.json does not exists, it will use a default config in it's place.
// Takes in an optional FlagOverride, will only use the first one passed, that
// can override the given flag values.
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
func LoadConfig(flags ...FlagOverride) (*Config, error) {
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
				UAAOAuthClient:       DefaultUAAOAuthClient,
				UAAOAuthClientSecret: DefaultUAAOAuthClientSecret,
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

		if config.ConfigFile.UAAOAuthClient == "" {
			config.ConfigFile.UAAOAuthClient = DefaultUAAOAuthClient
			config.ConfigFile.UAAOAuthClientSecret = DefaultUAAOAuthClientSecret
		}
	}

	config.ENV = EnvOverride{
		BinaryName:       filepath.Base(os.Args[0]),
		CFColor:          os.Getenv("CF_COLOR"),
		CFPluginHome:     os.Getenv("CF_PLUGIN_HOME"),
		CFStagingTimeout: os.Getenv("CF_STAGING_TIMEOUT"),
		CFStartupTimeout: os.Getenv("CF_STARTUP_TIMEOUT"),
		CFTrace:          os.Getenv("CF_TRACE"),
		HTTPSProxy:       os.Getenv("https_proxy"),
		Lang:             os.Getenv("LANG"),
		LCAll:            os.Getenv("LC_ALL"),
		Experimental:     os.Getenv("CF_CLI_EXPERIMENTAL"),
		CFDialTimeout:    os.Getenv("CF_DIAL_TIMEOUT"),
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

	if len(flags) > 0 {
		config.Flags = flags[0]
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

	// Flags stores the configuration from gobal flags
	Flags FlagOverride

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
	UAAOAuthClient           string        `json:"UAAOAuthClient"`
	UAAOAuthClientSecret     string        `json:"UAAOAuthClientSecret"`
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
	CFDialTimeout    string
}

// FlagOverride represents all the global flags passed to the CF CLI
type FlagOverride struct {
	Verbose bool
}

// Target returns the CC API URL
func (config *Config) Target() string {
	return config.ConfigFile.Target
}

// SkipSSLValidation returns whether or not to skip SSL validation when
// targeting an API endpoint
func (config *Config) SkipSSLValidation() bool {
	return config.ConfigFile.SkipSSLValidation
}

// AccessToken returns the access token for making authenticated API calls
func (config *Config) AccessToken() string {
	return config.ConfigFile.AccessToken
}

// RefreshToken returns the refresh token for getting a new access token
func (config *Config) RefreshToken() string {
	return config.ConfigFile.RefreshToken
}

// UAAOAuthClient returns the CLI's UAA client ID
func (config *Config) UAAOAuthClient() string {
	return config.ConfigFile.UAAOAuthClient
}

// UAAOAuthClientSecret returns the CLI's UAA client secret
func (config *Config) UAAOAuthClientSecret() string {
	return config.ConfigFile.UAAOAuthClientSecret
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

// Experimental returns whether or not to run experimental CLI commands. This
// is based off of:
//   1. The $CF_CLI_EXPERIMENTAL environment variable if set
//   2. Defaults to false
func (config *Config) Experimental() bool {
	if config.ENV.Experimental != "" {
		envVal, err := strconv.ParseBool(config.ENV.Experimental)
		if err == nil {
			return envVal
		}
	}

	return false
}

// Verbose returns true if verbose should be displayed to terminal and a
// location to log to. This is based off of:
//   - The config file's trace value (true/false/file path)
//   - The $CF_TRACE enviroment variable if set (true/false/file path)
//   - The '-v/--verbose' global flag
//   - Defaults to false
func (config *Config) Verbose() (bool, []string) {
	var (
		verbose     bool
		envOverride bool
		filePath    []string
	)
	if config.ENV.CFTrace != "" {
		envVal, err := strconv.ParseBool(config.ENV.CFTrace)
		verbose = envVal
		if err != nil {
			filePath = []string{config.ENV.CFTrace}
		} else {
			envOverride = true
		}
	}
	if config.ConfigFile.Trace != "" {
		envVal, err := strconv.ParseBool(config.ConfigFile.Trace)
		if !envOverride {
			verbose = envVal || verbose
		}
		if err != nil {
			filePath = append(filePath, config.ConfigFile.Trace)
		}
	}
	verbose = config.Flags.Verbose || verbose

	return verbose, filePath
}

// DialTimeout returns the timeout to use when dialing. This is based off of:
//   1. The $CF_DIAL_TIMEOUT environment variable if set
//   2. Defaults to 5 seconds
func (config *Config) DialTimeout() time.Duration {
	if config.ENV.CFDialTimeout != "" {
		envVal, err := strconv.ParseInt(config.ENV.CFDialTimeout, 10, 64)
		if err == nil {
			return time.Duration(envVal) * time.Second
		}
	}

	return DefaultDialTimeout
}

func (config *Config) BinaryVersion() string {
	return version.BinaryVersion
}

func (config *Config) BinaryBuildDate() string {
	return version.BinaryBuildDate
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

// SetAccessToken sets the current access token
func (config *Config) SetAccessToken(accessToken string) {
	config.ConfigFile.AccessToken = accessToken
}

// SetRefreshToken sets the current refresh token
func (config *Config) SetRefreshToken(refreshToken string) {
	config.ConfigFile.RefreshToken = refreshToken
}
