package clissh

import (
	"golang.org/x/crypto/ssh"
	"io"
)

//go:generate counterfeiter . SecureSession

type SecureSession interface {
	RequestPty(term string, height, width int, termModes ssh.TerminalModes) error
	SendRequest(name string, wantReply bool, payload []byte) (bool, error)
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	StderrPipe() (io.Reader, error)
	Start(command string) error
	Shell() error
	Wait() error
	Close() error
}
