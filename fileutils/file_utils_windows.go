package fileutils

import (
	"os"
	"syscall"
)

const (
	FILE_ATTRIBUTE_REPARSE_POINT = 0x0400
)

func IsRegular(f os.FileInfo) bool {
	if fileattrs, ok := f.Sys().(*syscall.Win32FileAttributeData); ok {
		if fileattrs.FileAttributes&FILE_ATTRIBUTE_REPARSE_POINT != 0 {
			return false
		}
	}
	return f.Mode().IsRegular()
}
