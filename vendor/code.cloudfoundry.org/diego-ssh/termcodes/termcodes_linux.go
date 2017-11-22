// +build linux

package termcodes

import (
	"os"
	"syscall"
	"unsafe"
)

type iflagSetter struct {
	Flag uint32
}

type lflagSetter struct {
	Flag uint32
}

type oflagSetter struct {
	Flag uint32
}

type cflagSetter struct {
	Flag uint32
}

func SetAttr(tty *os.File, termios *syscall.Termios) error {
	r, _, e := syscall.Syscall(syscall.SYS_IOCTL, tty.Fd(), syscall.TCSETS, uintptr(unsafe.Pointer(termios)))
	if r != 0 {
		return os.NewSyscallError("SYS_IOCTL", e)
	}

	return nil
}

func GetAttr(tty *os.File) (*syscall.Termios, error) {
	termios := &syscall.Termios{}

	r, _, e := syscall.Syscall(syscall.SYS_IOCTL, tty.Fd(), syscall.TCGETS, uintptr(unsafe.Pointer(termios)))
	if r != 0 {
		return nil, os.NewSyscallError("SYS_IOCTL", e)
	}

	return termios, nil
}
