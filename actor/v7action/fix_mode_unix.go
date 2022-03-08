//go:build !windows
// +build !windows

package v7action

import "os"

// fixMode is unnecessary on UNIX systems, see windows version for more
// details.
func fixMode(mode os.FileMode) os.FileMode {
	return mode
}
