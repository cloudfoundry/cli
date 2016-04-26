package confighelpers

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func DefaultFilePath() (string, error) {
	var homeDir string

	if os.Getenv("CF_HOME") != "" {
		homeDir = os.Getenv("CF_HOME")

		if _, err := os.Stat(homeDir); os.IsNotExist(err) {
			return "", fmt.Errorf("Error locating CF_HOME folder '%s'", homeDir)
		}
	} else {
		homeDir = userHomeDir()
	}

	return filepath.Join(homeDir, ".cf", "config.json"), nil
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
