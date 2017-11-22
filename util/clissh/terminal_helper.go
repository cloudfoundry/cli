package clissh

import (
	"io"

	"github.com/moby/moby/pkg/term"
)

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

func (terminalHelper) SetRawTerminal(fd uintptr) (*term.State, error) {
	return term.SetRawTerminal(fd)
}

func (terminalHelper) RestoreTerminal(fd uintptr, state *term.State) error {
	return term.RestoreTerminal(fd, state)
}

func (terminalHelper) StdStreams() (io.ReadCloser, io.Writer, io.Writer) {
	return term.StdStreams()
}
