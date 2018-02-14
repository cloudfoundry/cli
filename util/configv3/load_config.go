package configv3

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/command/translatableerror"
	"golang.org/x/crypto/ssh/terminal"
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
		ConfigFile: JSONConfig{
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
			// TODO: change this to not use translatableerror
			jsonError = translatableerror.EmptyConfigError{FilePath: configFilePath}
		} else {
			var configFile JSONConfig
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
