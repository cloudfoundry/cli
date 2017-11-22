// +build windows

package atime

import (
	"errors"
	"os"
	"syscall"
	"time"
)

func AccessTime(fileInfo os.FileInfo) (time.Time, error) {
	if fileInfo == nil || fileInfo.Sys() == nil {
		return time.Time{}, errors.New("underlying file information unavailable")
	}

	accessTime := fileInfo.Sys().(*syscall.Win32FileAttributeData).LastAccessTime.Nanoseconds()

	return time.Unix(int64(accessTime/1e9), int64(accessTime%1e9)), nil
}
