package config_helpers

import (
	"os"
	"path/filepath"
	"runtime"
)

func DefaultFilePath() string {
	var configDir string

	if os.Getenv("CF_HOME") != "" {
		cfHome := os.Getenv("CF_HOME")
		configDir = filepath.Join(cfHome, ".cf")
	} else {
		configDir = filepath.Join(userHomeDir(), ".cf")
	}

	return filepath.Join(configDir, "config.json")
}

// See: http://stackoverflow.com/questions/7922270/obtain-users-home-directory
// we can't cross compile using cgo and use user.Current()
var userHomeDir = func() string {

	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}

	return os.Getenv("HOME")
}

var PluginRepoDir = func() string {
	if os.Getenv("CF_PLUGIN_HOME") != "" {
		return os.Getenv("CF_PLUGIN_HOME")
	}

	return userHomeDir()
}
