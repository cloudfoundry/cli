// Package configv3 package contains everything related to the CF CLI Configuration.
package configv3

import (
	"path/filepath"
	"strconv"

	"code.cloudfoundry.org/cli/version"
)

// Config combines the settings taken from the .cf/config.json, os.ENV, and the
// plugin config.
type Config struct {
	// ConfigFile stores the configuration from the .cf/config
	ConfigFile JSONConfig

	// ENV stores the configuration from os.ENV
	ENV EnvOverride

	// Flags stores the configuration from gobal flags
	Flags FlagOverride

	// detectedSettings are settings detected when the config is loaded.
	detectedSettings detectedSettings

	pluginsConfig PluginsConfig
}

// BinaryVersion is the current version of the CF binary.
func (config *Config) BinaryVersion() string {
	return version.VersionString()
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

// TerminalWidth returns the width of the terminal from when the config
// was loaded. If the terminal width has changed since the config has loaded,
// it will **not** return the new width.
func (config *Config) TerminalWidth() int {
	return config.detectedSettings.terminalWidth
}

// Verbose returns true if verbose should be displayed to terminal, in addition
// a slice of absolute paths in which verbose text will appear. This is based
// off of:
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
		resolvedPath, err := filepath.EvalSymlinks(filePath[i])
		if err == nil {
			filePath[i] = resolvedPath
		}
	}

	return verbose, filePath
}
