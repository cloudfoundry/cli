// +build windows

package sigwinch

import "syscall"

func SIGWINCH() syscall.Signal {
	panic("Not supported on windows")
}
