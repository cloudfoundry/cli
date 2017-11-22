package handlers

import (
	"os/exec"
	"syscall"

	"code.cloudfoundry.org/lager"
	"golang.org/x/crypto/ssh"
)

//go:generate counterfeiter -o fake_handlers/fake_global_request_handler.go . GlobalRequestHandler
type GlobalRequestHandler interface {
	HandleRequest(logger lager.Logger, request *ssh.Request)
}

//go:generate counterfeiter -o fake_handlers/fake_new_channel_handler.go . NewChannelHandler
type NewChannelHandler interface {
	HandleNewChannel(logger lager.Logger, newChannel ssh.NewChannel)
}

//go:generate counterfeiter -o fakes/fake_runner.go . Runner
type Runner interface {
	Start(cmd *exec.Cmd) error
	Wait(cmd *exec.Cmd) error
	Signal(cmd *exec.Cmd, signal syscall.Signal) error
}

//go:generate counterfeiter -o fakes/fake_shell_locator.go . ShellLocator
type ShellLocator interface {
	ShellPath() string
}
