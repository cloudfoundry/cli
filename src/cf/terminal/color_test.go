package terminal

import (
	"github.com/stretchr/testify/assert"
	"os"
	"runtime"
	"testing"
)

func TestColorize(t *testing.T) {
	os.Setenv("CF_COLOR", "true")
	text := "Hello World"
	colorizedText := colorize(text, red, true)

	if runtime.GOOS == "windows" {
		assert.Equal(t, colorizedText, "Hello World")
	} else {
		assert.Equal(t, colorizedText, "\033[1;31mHello World\033[0m")
	}
}
