// +build !windows

package configv3

import (
	"os"
	"path/filepath"
)

// ConfigFilePath returns the location of the config file
func ConfigFilePath() string {
	return filepath.Join(HomeDirectory(true), ".cf", "config.json")
}

func HomeDirectory(checkCFHome bool) string {
	switch {
	case checkCFHome && os.Getenv("CF_HOME") != "":
		return os.Getenv("CF_HOME")
	default:
		return os.Getenv("HOME")
	}
}
