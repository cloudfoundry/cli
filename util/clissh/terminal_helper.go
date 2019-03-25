package clissh

import (
	"io"

	"github.com/moby/moby/pkg/term"
)

//go:generate counterfeiter . TerminalHelper

type TerminalHelper interface {
	GetFdInfo(in interface{}) (fd uintptr, isTerminal bool)
	SetRawTerminal(fd uintptr) (*term.State, error)
	RestoreTerminal(fd uintptr, state *term.State) error
	GetWinsize(fd uintptr) (*term.Winsize, error)
	StdStreams() (stdin io.ReadCloser, stdout io.Writer, stderr io.Writer)
}

type terminalHelper struct{}

func DefaultTerminalHelper() terminalHelper {
	return terminalHelper{}
}

func (terminalHelper) GetFdInfo(in interface{}) (uintptr, bool) {
	return term.GetFdInfo(in)
}

func (terminalHelper) GetWinsize(fd uintptr) (*term.Winsize, error) {
	return term.GetWinsize(fd)
}

func (terminalHelper) RestoreTerminal(fd uintptr, state *term.State) error {
	return term.RestoreTerminal(fd, state)
}

func (terminalHelper) SetRawTerminal(fd uintptr) (*term.State, error) {
	return term.SetRawTerminal(fd)
}

func (terminalHelper) StdStreams() (io.ReadCloser, io.Writer, io.Writer) {
	return term.StdStreams()
}
