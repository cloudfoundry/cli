// Package configv3 package contains everything related to the CF CLI Configuration.
package configv3

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/version"
)

const (
	// DefaultDialTimeout is the default timeout for the dail.
	DefaultDialTimeout = 5 * time.Second

	// DefaultNOAARetryCount is the default number of request retries.
	DefaultNOAARetryCount = 5

	// DefaultOverallPollingTimeout is the default maximum time that the CLI will
	// poll a job running on the Cloud Controller. By default it's infinit, which
	// is represented by MaxInt64.
	DefaultOverallPollingTimeout = time.Duration(1 << 62)
	// Developer Note: Due to bugs in using MaxInt64 during comparison, the above
	// was chosen as a replacement.

	// DefaultPollingInterval is the time between consecutive polls of a status.
	DefaultPollingInterval = 3 * time.Second

	// DefaultStagingTimeout is the default timeout for application staging.
	DefaultStagingTimeout = 15 * time.Minute

	// DefaultStartupTimeout is the default timeout for application starting.
	DefaultStartupTimeout = 5 * time.Minute
	// DefaultPingerThrottle = 5 * time.Second

	// DefaultTarget is the default CFConfig value for Target.
	DefaultTarget = ""

	// DefaultSSHOAuthClient is the default oauth client ID for SSHing into an
	// application/process container
	DefaultSSHOAuthClient = "ssh-proxy"

	// DefaultUAAOAuthClient is the default client ID for the CLI when
	// communicating with the UAA.
	DefaultUAAOAuthClient = "cf"

	// DefaultUAAOAuthClientSecret is the default client secret for the CLI when
	// communicating with the UAA.
	DefaultUAAOAuthClientSecret = ""

	// DefaultRetryCount is the default number of request retries.
	DefaultRetryCount = 2
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
	err := removeOldTempConfigFiles()
	if err != nil {
		return nil, err
	}

	configFilePath := ConfigFilePath()

	config := Config{
		ConfigFile: CFConfig{
			ConfigVersion: 3,
			Target:        DefaultTarget,
			ColorEnabled:  DefaultColorEnabled,
			PluginRepositories: []PluginRepository{{
				Name: DefaultPluginRepoName,
				URL:  DefaultPluginRepoURL,
			}},
		},
	}

	var jsonError error

	if _, err = os.Stat(configFilePath); err == nil || !os.IsNotExist(err) {
		var file []byte
		file, err = ioutil.ReadFile(configFilePath)
		if err != nil {
			return nil, err
		}

		if len(file) == 0 {
			jsonError = translatableerror.EmptyConfigError{FilePath: configFilePath}
		} else {
			var configFile CFConfig
			err = json.Unmarshal(file, &configFile)
			if err != nil {
				return nil, err
			}
			config.ConfigFile = configFile
		}
	}

	if config.ConfigFile.SSHOAuthClient == "" {
		config.ConfigFile.SSHOAuthClient = DefaultSSHOAuthClient
	}

	if config.ConfigFile.UAAOAuthClient == "" {
		config.ConfigFile.UAAOAuthClient = DefaultUAAOAuthClient
		config.ConfigFile.UAAOAuthClientSecret = DefaultUAAOAuthClientSecret
	}

	config.ENV = EnvOverride{
		BinaryName:       filepath.Base(os.Args[0]),
		CFColor:          os.Getenv("CF_COLOR"),
		CFDialTimeout:    os.Getenv("CF_DIAL_TIMEOUT"),
		CFLogLevel:       os.Getenv("CF_LOG_LEVEL"),
		CFPluginHome:     os.Getenv("CF_PLUGIN_HOME"),
		CFStagingTimeout: os.Getenv("CF_STAGING_TIMEOUT"),
		CFStartupTimeout: os.Getenv("CF_STARTUP_TIMEOUT"),
		CFTrace:          os.Getenv("CF_TRACE"),
		DockerPassword:   os.Getenv("CF_DOCKER_PASSWORD"),
		Experimental:     os.Getenv("CF_CLI_EXPERIMENTAL"),
		ForceTTY:         os.Getenv("FORCE_TTY"),
		HTTPSProxy:       os.Getenv("https_proxy"),
		Lang:             os.Getenv("LANG"),
		LCAll:            os.Getenv("LC_ALL"),
	}

	pluginFilePath := filepath.Join(config.PluginHome(), "config.json")
	if _, err = os.Stat(pluginFilePath); os.IsNotExist(err) {
		config.pluginsConfig = PluginsConfig{
			Plugins: make(map[string]Plugin),
		}
	} else {
		var file []byte
		file, err = ioutil.ReadFile(pluginFilePath)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(file, &config.pluginsConfig)
		if err != nil {
			return nil, err
		}

		for name, plugin := range config.pluginsConfig.Plugins {
			plugin.Name = name
			config.pluginsConfig.Plugins[name] = plugin
		}
	}

	if len(flags) > 0 {
		config.Flags = flags[0]
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Developer Note: The following is untested! Change at your own risk.
	isTTY := terminal.IsTerminal(int(os.Stdout.Fd()))
	terminalWidth := math.MaxInt32

	if isTTY {
		var err error
		terminalWidth, _, err = terminal.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			return nil, err
		}
	}

	config.detectedSettings = detectedSettings{
		currentDirectory: pwd,
		terminalWidth:    terminalWidth,
		tty:              isTTY,
	}

	return &config, jsonError
}

func removeOldTempConfigFiles() error {
	oldTempFileNames, err := filepath.Glob(filepath.Join(configDirectory(), "temp-config?*"))
	if err != nil {
		return err
	}

	for _, oldTempFileName := range oldTempFileNames {
		err = os.Remove(oldTempFileName)
		if err != nil {
			return err
		}
	}

	return nil
}

// WriteConfig creates the .cf directory and then writes the config.json. The
// location of .cf directory is written in the same way LoadConfig reads .cf
// directory.
func WriteConfig(c *Config) error {
	rawConfig, err := json.MarshalIndent(c.ConfigFile, "", "  ")
	if err != nil {
		return err
	}

	dir := configDirectory()
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	// Developer Note: The following is untested! Change at your own risk.
	// Setup notifications of termination signals to channel sig, create a process to
	// watch for these signals so we can remove transient config temp files.
	sig := make(chan os.Signal, 10)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(sig)

	tempConfigFile, err := ioutil.TempFile(dir, "temp-config")
	if err != nil {
		return err
	}
	tempConfigFile.Close()
	tempConfigFileName := tempConfigFile.Name()

	go catchSignal(sig, tempConfigFileName)

	err = ioutil.WriteFile(tempConfigFileName, rawConfig, 0600)
	if err != nil {
		return err
	}

	return os.Rename(tempConfigFileName, ConfigFilePath())
}

// catchSignal tries to catch SIGHUP, SIGINT, SIGKILL, SIGQUIT and SIGTERM, and
// Interrupt for removing temporarily created config files before the program
// ends.  Note:  we cannot intercept a `kill -9`, so a well-timed `kill -9`
// will allow a temp config file to linger.
func catchSignal(sig chan os.Signal, tempConfigFileName string) {
	select {
	case <-sig:
		_ = os.Remove(tempConfigFileName)
		os.Exit(2)
	}
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

	// detectedSettings are settings detected when the config is loaded.
	detectedSettings detectedSettings

	pluginsConfig PluginsConfig
}

// CFConfig represents .cf/config.json
type CFConfig struct {
	ConfigVersion            int                `json:"ConfigVersion"`
	Target                   string             `json:"Target"`
	APIVersion               string             `json:"APIVersion"`
	AuthorizationEndpoint    string             `json:"AuthorizationEndpoint"`
	DopplerEndpoint          string             `json:"DopplerEndPoint"`
	UAAEndpoint              string             `json:"UaaEndpoint"`
	RoutingEndpoint          string             `json:"RoutingAPIEndpoint"`
	AccessToken              string             `json:"AccessToken"`
	SSHOAuthClient           string             `json:"SSHOAuthClient"`
	UAAOAuthClient           string             `json:"UAAOAuthClient"`
	UAAOAuthClientSecret     string             `json:"UAAOAuthClientSecret"`
	RefreshToken             string             `json:"RefreshToken"`
	TargetedOrganization     Organization       `json:"OrganizationFields"`
	TargetedSpace            Space              `json:"SpaceFields"`
	SkipSSLValidation        bool               `json:"SSLDisabled"`
	AsyncTimeout             int                `json:"AsyncTimeout"`
	Trace                    string             `json:"Trace"`
	ColorEnabled             string             `json:"ColorEnabled"`
	Locale                   string             `json:"Locale"`
	PluginRepositories       []PluginRepository `json:"PluginRepos"`
	MinCLIVersion            string             `json:"MinCLIVersion"`
	MinRecommendedCLIVersion string             `json:"MinRecommendedCLIVersion"`
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
	CFDialTimeout    string
	CFHome           string
	CFLogLevel       string
	CFPluginHome     string
	CFStagingTimeout string
	CFStartupTimeout string
	CFTrace          string
	DockerPassword   string
	Experimental     string
	ForceTTY         string
	HTTPSProxy       string
	Lang             string
	LCAll            string
}

// FlagOverride represents all the global flags passed to the CF CLI
type FlagOverride struct {
	Verbose bool
}

// detectedSettings are automatically detected settings determined by the CLI.
type detectedSettings struct {
	currentDirectory string
	terminalWidth    int
	tty              bool
}

// Target returns the CC API URL
func (config *Config) Target() string {
	return config.ConfigFile.Target
}

// PollingInterval returns the time between polls.
func (config *Config) PollingInterval() time.Duration {
	return DefaultPollingInterval
}

// OverallPollingTimeout returns the overall polling timeout for async
// operations. The time is based off of:
//   1. The config file's AsyncTimeout value (integer) is > 0
//   2. Defaults to the DefaultOverallPollingTimeout
func (config *Config) OverallPollingTimeout() time.Duration {
	if config.ConfigFile.AsyncTimeout == 0 {
		return DefaultOverallPollingTimeout
	}
	return time.Duration(config.ConfigFile.AsyncTimeout) * time.Minute
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

// SSHOAuthClient returns the OAuth client id used for SSHing into
// application/process containers
func (config *Config) SSHOAuthClient() string {
	return config.ConfigFile.SSHOAuthClient
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

// MinCLIVersion returns the minimum CLI version requried by the CC
func (config *Config) MinCLIVersion() string {
	return config.ConfigFile.MinCLIVersion
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

// Verbose returns true if verbose should be displayed to terminal, in addition
// a slice of full paths in which verbose text will appear. This is based off
// of:
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

	for i, path := range filePath {
		if !filepath.IsAbs(path) {
			filePath[i] = filepath.Join(config.detectedSettings.currentDirectory, path)
		}
	}

	return verbose, filePath
}

// IsTTY returns true based off of:
//   - The $FORCE_TTY is set to true/t/1
//   - Detected from the STDOUT stream
func (config *Config) IsTTY() bool {
	if config.ENV.ForceTTY != "" {
		envVal, err := strconv.ParseBool(config.ENV.ForceTTY)
		if err == nil {
			return envVal
		}
	}

	return config.detectedSettings.tty
}

// LogLevel returns the global log level. The levels follow Logrus's log level
// scheme. This value is based off of:
//   - The $CF_LOG_LEVEL and an int/warn/info/etc...
//   - Defaults to PANIC/0 (ie no logging)
func (config *Config) LogLevel() int {
	if config.ENV.CFLogLevel != "" {
		envVal, err := strconv.ParseInt(config.ENV.CFLogLevel, 10, 32)
		if err == nil {
			return int(envVal)
		}

		switch strings.ToLower(config.ENV.CFLogLevel) {
		case "fatal":
			return 1
		case "error":
			return 2
		case "warn":
			return 3
		case "info":
			return 4
		case "debug":
			return 5
		}
	}

	return 0
}

// TerminalWidth returns the width of the terminal from when the config
// was loaded. If the terminal width has changed since the config has loaded,
// it will **not** return the new width.
func (config *Config) TerminalWidth() int {
	return config.detectedSettings.terminalWidth
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
	return version.VersionString()
}

// HasTargetedOrganization returns true if the organization is set
func (config *Config) HasTargetedOrganization() bool {
	return config.ConfigFile.TargetedOrganization.GUID != ""
}

// HasTargetedSpace returns true if the space is set
func (config *Config) HasTargetedSpace() bool {
	return config.ConfigFile.TargetedSpace.GUID != ""
}

// DockerPassword returns the docker password from the environment.
func (config *Config) DockerPassword() string {
	return config.ENV.DockerPassword
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
func (config *Config) SetTargetInformation(api string, apiVersion string, auth string, minCLIVersion string, doppler string, routing string, skipSSLValidation bool) {
	config.ConfigFile.Target = api
	config.ConfigFile.APIVersion = apiVersion
	config.ConfigFile.AuthorizationEndpoint = auth
	config.ConfigFile.MinCLIVersion = minCLIVersion
	config.ConfigFile.DopplerEndpoint = doppler
	config.ConfigFile.RoutingEndpoint = routing
	config.ConfigFile.SkipSSLValidation = skipSSLValidation

	config.UnsetOrganizationInformation()
	config.UnsetSpaceInformation()
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

// SetUAAEndpoint sets the UAA endpoint that is obtained from hitting
// <AuthorizationEndpoint>/login
func (config *Config) SetUAAEndpoint(uaaEndpoint string) {
	config.ConfigFile.UAAEndpoint = uaaEndpoint
}

// UnsetSpaceInformation resets the space values to default
func (config *Config) UnsetSpaceInformation() {
	config.SetSpaceInformation("", "", false)
}

// UnsetOrganizationInformation resets the organization values to default
func (config *Config) UnsetOrganizationInformation() {
	config.SetOrganizationInformation("", "")

}

// RequestRetryCount returns the number of request retries.
func (*Config) RequestRetryCount() int {
	return DefaultRetryCount
}

// NOAARequestRetryCount returns the number of request retries.
func (*Config) NOAARequestRetryCount() int {
	return DefaultNOAARetryCount
}
