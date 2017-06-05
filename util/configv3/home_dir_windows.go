// +build windows

package configv3

import (
	"os"
	"path/filepath"
)

func ConfigFilePath() string {
	return filepath.Join(HomeDirectory(false), ".cf", "config.json")
}

// See: http://stackoverflow.com/questions/7922270/obtain-users-home-directory
// we can't cross compile using cgo and use user.Current()
func HomeDirectory(checkCFHome bool) string {
	switch {
	case checkCFHome && os.Getenv("CF_HOME") != "":
		return os.Getenv("CF_HOME")
	case os.Getenv("HOMEDRIVE")+os.Getenv("HOMEPATH") != "":
		return os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	default:
		return os.Getenv("USERPROFILE")
	}
}
