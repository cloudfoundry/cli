// +build linux

package atime

import (
	"os"
	"syscall"
)

func accessTimespec(fileInfo os.FileInfo) syscall.Timespec {
	return fileInfo.Sys().(*syscall.Stat_t).Atim
}
