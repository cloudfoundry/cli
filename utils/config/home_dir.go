// +build !windows

package config

import (
	"os"
	"path/filepath"
)

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

// DefaultFilePath returns the location of the config file
func defaultFilePath() string {
	return filepath.Join(homeDirectory(), ".cf", "config.json")
}
