package fileutil

import (
	"os"
	"runtime"
	"syscall"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type fileMover struct {
	fs boshsys.FileSystem
}

func NewFileMover(fs boshsys.FileSystem) fileMover {
	return fileMover{fs: fs}
}

func (m fileMover) Move(oldPath, newPath string) error {
	err := m.fs.Rename(oldPath, newPath)

	le, ok := err.(*os.LinkError)
	if !ok {
		return err
	}

	// 0x11 is Win32 Error Code ERROR_NOT_SAME_DEVICE (https://msdn.microsoft.com/en-us/library/cc231199.aspx)
	if le.Err == syscall.Errno(0x12) || (runtime.GOOS == "windows" && le.Err == syscall.Errno(0x11)) {
		err = m.fs.CopyFile(oldPath, newPath)
		if err != nil {
			return err
		}

		err = m.fs.RemoveAll(oldPath)
		if err != nil {
			return err
		}

		return nil
	}

	return err
}
