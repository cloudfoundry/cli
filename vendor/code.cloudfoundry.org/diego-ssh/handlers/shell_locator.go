package handlers

import "os/exec"

type shellLocator struct{}

func NewShellLocator() ShellLocator {
	return &shellLocator{}
}

func (shellLocator) ShellPath() string {
	for _, shell := range []string{"/bin/bash", "/usr/local/bin/bash", "/bin/sh", "bash", "sh"} {
		if path, err := exec.LookPath(shell); err == nil {
			return path
		}
	}

	return "/bin/sh"
}
