//go:build !windows
// +build !windows

package configv3

import (
	"os"
	"path/filepath"
)

// ConfigFilePath returns the location of the config file
func ConfigFilePath() string {
	return filepath.Join(configDirectory(), "config.json")
}

func configDirectory() string {
	return filepath.Join(homeDirectory(), ".cf")
}

func homeDirectory() string {
	var homeDir string
	switch {
	case os.Getenv("CF_HOME") != "":
		homeDir = os.Getenv("CF_HOME")
	default:
		homeDir = os.Getenv("HOME")
	}
	return homeDir
}
