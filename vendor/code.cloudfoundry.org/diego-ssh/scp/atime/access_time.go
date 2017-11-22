// +build !windows

package atime

import (
	"errors"
	"os"
	"time"
)

func AccessTime(fileInfo os.FileInfo) (time.Time, error) {
	if fileInfo == nil || fileInfo.Sys() == nil {
		return time.Time{}, errors.New("underlying file information unavailable")
	}

	timespec := accessTimespec(fileInfo)

	return time.Unix(int64(timespec.Sec), int64(timespec.Nsec)), nil
}
