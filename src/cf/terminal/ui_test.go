package terminal

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"io"
	"os"
)

func simulateStdin(input string, block func()) {
	defer func() {
		stdin = os.Stdin
	}()

	stdinReader, stdinWriter := io.Pipe()
	stdin = stdinReader

	go func() {
		stdinWriter.Write([]byte(input))
		defer stdinWriter.Close()
	}()

	block()
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = old
	return <-outC
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSayWithStringOnly", func() {
			ui := new(terminalUI)
			out := captureOutput(func() {
				ui.Say("Hello")
			})

			assert.Equal(mr.T(), "Hello\n", out)
		})
		It("TestSayWithStringWithFormat", func() {

			ui := new(terminalUI)
			out := captureOutput(func() {
				ui.Say("Hello %s", "World!")
			})

			assert.Equal(mr.T(), "Hello World!\n", out)
		})
		It("TestConfirmYes", func() {

			simulateStdin("y\n", func() {
				ui := new(terminalUI)

				var result bool
				out := captureOutput(func() {
					result = ui.Confirm("Hello %s", "World?")
				})

				assert.True(mr.T(), result)
				assert.Contains(mr.T(), out, "Hello World?")
			})
		})
		It("TestConfirmNo", func() {

			simulateStdin("wat\n", func() {
				ui := new(terminalUI)

				var result bool
				out := captureOutput(func() {
					result = ui.Confirm("Hello %s", "World?")
				})

				assert.False(mr.T(), result)
				assert.Contains(mr.T(), out, "Hello World?")
			})
		})
	})
}
