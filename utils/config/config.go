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

type Config struct {
	ConfigFile   CFConfig
	ENV          EnvOverride
	pluginConfig PluginsConfig
}

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

type Organization struct {
	GUID            string          `json:"GUID"`
	Name            string          `json:"Name"`
	QuotaDefinition QuotaDefinition `json:"QuotaDefinition"`
}

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

type Space struct {
	GUID     string `json:"GUID"`
	Name     string `json:"Name"`
	AllowSSH bool   `json:"AllowSSH"`
}

type PluginRepos struct {
	Name string `json:"Name"`
	URL  string `json:"URL"`
}

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
}

type PluginsConfig struct {
	Plugins map[string]Plugin `json:"Plugins"`
}

type Plugin struct {
	Location string         `json:"Location"`
	Version  PluginVersion  `json:"Version"`
	Commands PluginCommands `json:"Commands"`
}

type PluginVersion struct {
	Major int `json:"Major"`
	Minor int `json:"Minor"`
	Build int `json:"Build"`
}

type PluginCommands []PluginCommand

func (p PluginCommands) Len() int               { return len(p) }
func (p PluginCommands) Swap(i int, j int)      { p[i], p[j] = p[j], p[i] }
func (p PluginCommands) Less(i int, j int) bool { return sortutils.SortAlphabetic(p[i].Name, p[j].Name) }

type PluginCommand struct {
	Name         string             `json:"Name"`
	Alias        string             `json:"Alias"`
	HelpText     string             `json:"HelpText"`
	UsageDetails PluginUsageDetails `json:"UsageDetails"`
}

type PluginUsageDetails struct {
	Usage   string            `json:"Usage"`
	Options map[string]string `json:"Options"`
}

type ColorSetting int

const (
	ColorDisbled ColorSetting = iota
	ColorEnabled
	ColorAuto
)

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

func (_ *Config) boolToColorSetting(val bool) ColorSetting {
	if val {
		return ColorEnabled
	}

	return ColorDisbled
}

func (conf *Config) Target() string {
	return conf.ConfigFile.Target
}

func (conf *Config) APIVersion() string {
	return conf.ConfigFile.APIVersion
}

func (conf *Config) TargetedOrganization() Organization {
	return conf.ConfigFile.TargetedOrganization
}

func (conf *Config) TargetedSpace() Space {
	return conf.ConfigFile.TargetedSpace
}

func (conf *Config) CurrentUser() (User, error) {
	return decodeUserFromJWT(conf.ConfigFile.AccessToken)
}

func (config *Config) PluginHome() string {
	if config.ENV.CFPluginHome != "" {
		return filepath.Join(config.ENV.CFPluginHome, ".cf", "plugins")
	}

	return filepath.Join(homeDirectory(), ".cf", "plugins")
}

func (config *Config) Plugins() map[string]Plugin {
	return config.pluginConfig.Plugins
}

func (config *Config) StagingTimeout() time.Duration {
	if config.ENV.CFStagingTimeout != "" {
		val, err := strconv.ParseInt(config.ENV.CFStagingTimeout, 10, 64)
		if err == nil {
			return time.Duration(val) * time.Minute
		}
	}

	return DefaultStagingTimeout
}

func (config *Config) StartupTimeout() time.Duration {
	if config.ENV.CFStartupTimeout != "" {
		val, err := strconv.ParseInt(config.ENV.CFStartupTimeout, 10, 64)
		if err == nil {
			return time.Duration(val) * time.Minute
		}
	}

	return DefaultStartupTimeout
}

func (config *Config) HTTPSProxy() string {
	if config.ENV.HTTPSProxy != "" {
		return config.ENV.HTTPSProxy
	}

	return ""
}

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

func (config *Config) BinaryName() string {
	return config.ENV.BinaryName
}

func (config *Config) SetOrganizationInformation(guid string, name string) {
	config.ConfigFile.TargetedOrganization.GUID = guid
	config.ConfigFile.TargetedOrganization.Name = name
	config.ConfigFile.TargetedOrganization.QuotaDefinition = QuotaDefinition{}
}

func (config *Config) SetSpaceInformation(guid string, name string, allowSSH bool) {
	config.ConfigFile.TargetedSpace.GUID = guid
	config.ConfigFile.TargetedSpace.Name = name
	config.ConfigFile.TargetedSpace.AllowSSH = allowSSH
}

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

func (config *Config) SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string) {
	config.ConfigFile.AccessToken = accessToken
	config.ConfigFile.RefreshToken = refreshToken
	config.ConfigFile.SSHOAuthClient = sshOAuthClient
}

func (config *Config) PluginRepos() []PluginRepos {
	return config.ConfigFile.PluginRepos
}

func (config *Config) convertLocale(local string) string {
	lang := strings.Split(local, ".")[0]
	return strings.Replace(lang, "_", "-", -1)
}
