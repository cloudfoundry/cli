// +build darwin

package atime

import (
	"os"
	"syscall"
)

func accessTimespec(fileInfo os.FileInfo) syscall.Timespec {
	return fileInfo.Sys().(*syscall.Stat_t).Atimespec
}
