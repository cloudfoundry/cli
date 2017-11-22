package handlers

import (
	"os/exec"
	"syscall"
)

type commandRunner struct{}

func NewCommandRunner() Runner {
	return &commandRunner{}
}

func (commandRunner) Start(cmd *exec.Cmd) error {
	return cmd.Start()
}

func (commandRunner) Wait(cmd *exec.Cmd) error {
	return cmd.Wait()
}

func (commandRunner) Signal(cmd *exec.Cmd, signal syscall.Signal) error {
	return cmd.Process.Signal(signal)
}
