package terminalcolor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestColorize(t *testing.T) {
	text := "Hello World"
	colorizedText := Colorize(text, Blue, true)

	assert.Equal(t, colorizedText, "\033[1;34mHello World\033[0m")
}
