// +build !windows

package v2action

import "os"

func fixMode(mode os.FileMode) os.FileMode {
	return mode
}
