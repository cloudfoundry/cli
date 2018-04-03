package ui

import (
	"io"
	"strings"
	"sync"
)

type ComboWriter struct {
	ui        UI
	uiLock    sync.Mutex
	onNewLine bool
}

type prefixedWriter struct {
	w      *ComboWriter
	prefix string
}

func NewComboWriter(ui UI) *ComboWriter {
	return &ComboWriter{ui: ui, onNewLine: true}
}

func (w *ComboWriter) Writer(prefix string) io.Writer {
	return prefixedWriter{w: w, prefix: prefix}
}

func (s prefixedWriter) Write(bytes []byte) (int, error) {
	if len(bytes) == 0 {
		return 0, nil
	}

	s.w.uiLock.Lock()
	defer s.w.uiLock.Unlock()

	lines := strings.Split(string(bytes), "\n")

	for i, line := range lines {
		lastLine := i == len(lines)-1

		if !lastLine || len(line) > 0 {
			if s.w.onNewLine {
				s.w.ui.PrintBlock([]byte(s.prefix))
			}

			s.w.ui.PrintBlock([]byte(line))
			s.w.onNewLine = false

			if !lastLine {
				s.w.ui.PrintBlock([]byte("\n"))
				s.w.onNewLine = true
			}
		}
	}

	return len(bytes), nil
}
