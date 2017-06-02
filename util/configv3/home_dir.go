// +build !windows

package configv3

import (
	"os"
	"path/filepath"
)

// ConfigFilePath returns the location of the config file
func ConfigFilePath() string {
	return filepath.Join(homeDirectory(true), ".cf", "config.json")
}

func homeDirectory(checkCFHome bool) string {
	var homeDir string
	switch {
	case checkCFHome && os.Getenv("CF_HOME") != "":
		homeDir = os.Getenv("CF_HOME")
	default:
		homeDir = os.Getenv("HOME")
	}
	return homeDir
}
