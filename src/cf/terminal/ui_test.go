package terminal

import (
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestSayWithStringOnly(t *testing.T) {
	ui := new(TerminalUI)
	out := testhelpers.CaptureOutput(func() {
		ui.Say("Hello")
	})

	assert.Equal(t, "Hello\n", out)
}

func TestSayWithStringWithFormat(t *testing.T) {
	ui := new(TerminalUI)
	out := testhelpers.CaptureOutput(func() {
		ui.Say("Hello %s", "World!")
	})

	assert.Equal(t, "Hello World!\n", out)
}
