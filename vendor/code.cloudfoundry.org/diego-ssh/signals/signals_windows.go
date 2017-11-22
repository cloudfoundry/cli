// +build windows

package signals

import (
	"syscall"

	"golang.org/x/crypto/ssh"
)

var SyscallSignals = map[ssh.Signal]syscall.Signal{
	ssh.SIGABRT: syscall.SIGABRT,
	ssh.SIGALRM: syscall.SIGALRM,
	ssh.SIGFPE:  syscall.SIGFPE,
	ssh.SIGHUP:  syscall.SIGHUP,
	ssh.SIGILL:  syscall.SIGILL,
	ssh.SIGINT:  syscall.SIGINT,
	ssh.SIGKILL: syscall.SIGKILL,
	ssh.SIGPIPE: syscall.SIGPIPE,
	ssh.SIGQUIT: syscall.SIGQUIT,
	ssh.SIGSEGV: syscall.SIGSEGV,
	ssh.SIGTERM: syscall.SIGTERM,
}

var SSHSignals = map[syscall.Signal]ssh.Signal{
	syscall.SIGABRT: ssh.SIGABRT,
	syscall.SIGALRM: ssh.SIGALRM,
	syscall.SIGFPE:  ssh.SIGFPE,
	syscall.SIGHUP:  ssh.SIGHUP,
	syscall.SIGILL:  ssh.SIGILL,
	syscall.SIGINT:  ssh.SIGINT,
	syscall.SIGKILL: ssh.SIGKILL,
	syscall.SIGPIPE: ssh.SIGPIPE,
	syscall.SIGQUIT: ssh.SIGQUIT,
	syscall.SIGSEGV: ssh.SIGSEGV,
	syscall.SIGTERM: ssh.SIGTERM,
}
