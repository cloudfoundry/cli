package terminal_test

import (
	"cf/terminal"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	testterm "testhelpers/terminal"
	"testing"
)

func TestSayWithStringOnly(t *testing.T) {
	ui := new(terminal.TerminalUI)
	out := testterm.CaptureOutput(func() {
		ui.Say("Hello")
	})

	assert.Equal(t, "Hello\n", out)
}

func TestSayWithStringWithFormat(t *testing.T) {
	ui := new(terminal.TerminalUI)
	out := testterm.CaptureOutput(func() {
		ui.Say("Hello %s", "World!")
	})

	assert.Equal(t, "Hello World!\n", out)
}

func TestConfirmYes(t *testing.T) {
	simulateStdin("y\n", func() {
		ui := new(terminal.TerminalUI)

		var result bool
		out := testterm.CaptureOutput(func() {
			result = ui.Confirm("Hello %s", "World?")
		})

		assert.True(t, result)
		assert.Contains(t, out, "Hello World?")
	})
}

func TestConfirmNo(t *testing.T) {
	simulateStdin("wat\n", func() {
		ui := new(terminal.TerminalUI)

		var result bool
		out := testterm.CaptureOutput(func() {
			result = ui.Confirm("Hello %s", "World?")
		})

		assert.False(t, result)
		assert.Contains(t, out, "Hello World?")
	})
}

func simulateStdin(input string, block func()) {
	defer func() {
		terminal.Stdin = os.Stdin
	}()

	stdinReader, stdinWriter := io.Pipe()
	terminal.Stdin = stdinReader

	go func() {
		stdinWriter.Write([]byte(input))
		defer stdinWriter.Close()
	}()

	block()
}
