package terminal

import (
	"io"

	"github.com/moby/moby/pkg/term"
)

//go:generate counterfeiter . TerminalHelper

type TerminalHelper interface {
	StdStreams() (stdin io.ReadCloser, stdout io.Writer, stderr io.Writer)
	GetFdInfo(in interface{}) (fd uintptr, isTerminal bool)
	SetRawTerminal(fd uintptr) (*term.State, error)
	RestoreTerminal(fd uintptr, state *term.State) error
	IsTerminal(fd uintptr) bool
	GetWinsize(fd uintptr) (*term.Winsize, error)
}

type terminalHelper struct{}

func DefaultHelper() TerminalHelper {
	return &terminalHelper{}
}

func (t *terminalHelper) StdStreams() (io.ReadCloser, io.Writer, io.Writer) {
	return term.StdStreams()
}

func (t *terminalHelper) GetFdInfo(in interface{}) (uintptr, bool) {
	return term.GetFdInfo(in)
}

func (t *terminalHelper) SetRawTerminal(fd uintptr) (*term.State, error) {
	return term.SetRawTerminal(fd)
}

func (t *terminalHelper) RestoreTerminal(fd uintptr, state *term.State) error {
	return term.RestoreTerminal(fd, state)
}

func (t *terminalHelper) IsTerminal(fd uintptr) bool {
	return term.IsTerminal(fd)
}

func (t *terminalHelper) GetWinsize(fd uintptr) (*term.Winsize, error) {
	return term.GetWinsize(fd)
}
