// +build windows

package sharedaction

import "os"

// fixMode forces all files on windows to be executable because by default
// everything on windows is read/write only. Even executable files.
func fixMode(mode os.FileMode) os.FileMode {
	return mode | 0700
}
