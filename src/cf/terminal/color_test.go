package terminal

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"os"
	"runtime"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestColorize", func() {

			os.Setenv("CF_COLOR", "true")
			text := "Hello World"
			colorizedText := colorize(text, red, true)

			if runtime.GOOS == "windows" {
				assert.Equal(mr.T(), colorizedText, "Hello World")
			} else {
				assert.Equal(mr.T(), colorizedText, "\033[1;31mHello World\033[0m")
			}
		})
	})
}
