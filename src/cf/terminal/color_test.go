package terminal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestColorize(t *testing.T) {
	text := "Hello World"
	colorizedText := colorize(text, red, true)

	assert.Equal(t, colorizedText, "\033[1;31mHello World\033[0m")
}
