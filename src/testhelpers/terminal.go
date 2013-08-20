package testhelpers

import (
	"bytes"
	"io"
	"os"
	"fmt"
)

func CaptureOutput(f func()) string {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	return <-outC
}

type FakeUI struct {
	Outputs []string
	Prompts []string
	Inputs []string
}

func (c *FakeUI) Say(message string, args ...interface{}) {
	c.Outputs = append(c.Outputs, fmt.Sprintf(message, args...))
	return
}

func (c *FakeUI) Ask(prompt string, args ...interface{}) (answer string) {
	c.Prompts = append(c.Prompts, fmt.Sprintf(prompt, args...))
	answer = c.Inputs[0]
	c.Inputs = c.Inputs[1:]
	return
}

func (c *FakeUI) Ok() {
	c.Say("OK")
}

func (c *FakeUI) Failed(message string, err error) {
	c.Say("FAILED")

	if message != "" {
		c.Say(message)
	}

	if err != nil {
		c.Say(err.Error())
	}
	return
}
