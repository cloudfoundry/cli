//go:build !windows
// +build !windows

package sigwinch

import "syscall"

func SIGWINCH() syscall.Signal {
	return syscall.SIGWINCH
}
