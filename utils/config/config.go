package config

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

	DefaultTarget       = "https://api.bosh-lite.com"
	DefaultColorEnabled = "true"
)

func LoadConfig() (*Config, error) {
	filePath := defaultFilePath()

	var config Config
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		config = Config{
			configFile: CFConfig{
				Target:       DefaultTarget,
				ColorEnabled: DefaultColorEnabled,
			},
		}
	} else {
		file, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(file, &config.configFile)
		if err != nil {
			return nil, err
		}
	}

	config.env = EnvOverride{
		CFColor:          os.Getenv("CF_COLOR"),
		CFPluginHome:     os.Getenv("CF_PLUGIN_HOME"),
		CFStagingTimeout: os.Getenv("CF_STAGING_TIMEOUT"),
		CFStartupTimeout: os.Getenv("CF_STARTUP_TIMEOUT"),
		CFTrace:          os.Getenv("CF_TRACE"),
		HTTPSProxy:       os.Getenv("https_proxy"),
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

type Config struct {
	configFile   CFConfig
	env          EnvOverride
	pluginConfig PluginsConfig
}

type CFConfig struct {
	ConfigVersion            int                `json:"ConfigVersion"`
	Target                   string             `json:"Target"`
	APIVersion               string             `json:"APIVersion"`
	AuthorizationEndpoint    string             `json:"AuthorizationEndpoint"`
	LoggregatorEndPoint      string             `json:"LoggregatorEndPoint"`
	DopplerEndPoint          string             `json:"DopplerEndPoint"`
	UaaEndpoint              string             `json:"UaaEndpoint"`
	RoutingAPIEndpoint       string             `json:"RoutingAPIEndpoint"`
	AccessToken              string             `json:"AccessToken"`
	SSHOAuthClient           string             `json:"SSHOAuthClient"`
	RefreshToken             string             `json:"RefreshToken"`
	OrganizationFields       OrganizationFields `json:"OrganizationFields"`
	SpaceFields              SpaceFields        `json:"SpaceFields"`
	SSLDisabled              bool               `json:"SSLDisabled"`
	AsyncTimeout             int                `json:"AsyncTimeout"`
	Trace                    string             `json:"Trace"`
	ColorEnabled             string             `json:"ColorEnabled"`
	Locale                   string             `json:"Locale"`
	PluginRepos              []PluginRepos      `json:"PluginRepos"`
	MinCLIVersion            string             `json:"MinCLIVersion"`
	MinRecommendedCLIVersion string             `json:"MinRecommendedCLIVersion"`
}

type OrganizationFields struct {
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

type SpaceFields struct {
	GUID     string `json:"GUID"`
	Name     string `json:"Name"`
	AllowSSH bool   `json:"AllowSSH"`
}

type PluginRepos struct {
	Name string `json:"Name"`
	URL  string `json:"URL"`
}

type EnvOverride struct {
	CFColor          string
	CFHome           string
	CFPluginHome     string
	CFStagingTimeout string
	CFStartupTimeout string
	CFTrace          string
	HTTPSProxy       string
}

type PluginsConfig struct {
	Plugins map[string]PluginConfig `json:"Plugins"`
}

type PluginConfig struct {
	Location string          `json:"Location"`
	Version  PluginVersion   `json:"Version"`
	Commands []PluginCommand `json:"Commands"`
}

type PluginVersion struct {
	Major int `json:"Major"`
	Minor int `json:"Minor"`
	Build int `json:"Build"`
}

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
	if config.env.CFColor != "" {
		val, err := strconv.ParseBool(config.env.CFColor)
		if err == nil {
			return config.boolToColorSetting(val)
		}
	}

	val, err := strconv.ParseBool(config.configFile.ColorEnabled)
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
	return conf.configFile.Target
}

func (config *Config) PluginHome() string {
	if config.env.CFPluginHome != "" {
		return filepath.Join(config.env.CFPluginHome, ".cf", "plugins")
	}

	return filepath.Join(homeDirectory(), ".cf", "plugins")
}

func (config *Config) PluginConfig() map[string]PluginConfig {
	return config.pluginConfig.Plugins
}

func (config *Config) StagingTimeout() time.Duration {
	if config.env.CFStagingTimeout != "" {
		val, err := strconv.ParseInt(config.env.CFStagingTimeout, 10, 64)
		if err == nil {
			return time.Duration(val) * time.Minute
		}
	}

	return DefaultStagingTimeout
}

func (config *Config) StartupTimeout() time.Duration {
	if config.env.CFStartupTimeout != "" {
		val, err := strconv.ParseInt(config.env.CFStartupTimeout, 10, 64)
		if err == nil {
			return time.Duration(val) * time.Minute
		}
	}

	return DefaultStartupTimeout
}

func (config Config) HTTPSProxy() string {
	if config.env.HTTPSProxy != "" {
		return config.env.HTTPSProxy
	}

	return ""
}
