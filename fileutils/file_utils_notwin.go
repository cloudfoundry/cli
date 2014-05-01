//

// +build !windows

package fileutils

import (
	"os"
)

func IsRegular(f os.FileInfo) bool {
	return f.Mode().IsRegular()
}
