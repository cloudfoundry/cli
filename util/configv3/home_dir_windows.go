//go:build windows
// +build windows

package configv3

import (
	"os"
	"path/filepath"
)

// ConfigFilePath returns the location of the config file
func ConfigFilePath() string {
	return filepath.Join(homeDirectory(), ".cf", "config.json")
}

func configDirectory() string {
	return filepath.Join(homeDirectory(), ".cf")
}

// See: http://stackoverflow.com/questions/7922270/obtain-users-home-directory
// we can't cross compile using cgo and use user.Current()
func homeDirectory() string {
	var homeDir string
	switch {
	case os.Getenv("CF_HOME") != "":
		homeDir = os.Getenv("CF_HOME")
	case os.Getenv("HOMEDRIVE")+os.Getenv("HOMEPATH") != "":
		homeDir = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	default:
		homeDir = os.Getenv("USERPROFILE")
	}
	return homeDir
}
