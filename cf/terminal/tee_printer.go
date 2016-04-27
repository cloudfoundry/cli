package terminal

import (
	"fmt"
	"io"
	"io/ioutil"
)

type TeePrinter struct {
	disableTerminalOutput bool
	outputBucket          io.Writer
}

func NewTeePrinter() *TeePrinter {
	return &TeePrinter{
		outputBucket: ioutil.Discard,
	}
}

func (t *TeePrinter) SetOutputBucket(bucket io.Writer) {
	if bucket == nil {
		bucket = ioutil.Discard
	}

	t.outputBucket = bucket
}

func (t *TeePrinter) Print(values ...interface{}) (int, error) {
	str := fmt.Sprint(values...)
	t.saveOutputToBucket(str)
	if !t.disableTerminalOutput {
		return PrintToTerminal(str)
	}
	return 0, nil
}

func (t *TeePrinter) Printf(format string, a ...interface{}) (int, error) {
	str := fmt.Sprintf(format, a...)
	t.saveOutputToBucket(str)
	if !t.disableTerminalOutput {
		return PrintToTerminal(str)
	}
	return 0, nil
}

func (t *TeePrinter) Println(values ...interface{}) (int, error) {
	str := fmt.Sprint(values...)
	t.saveOutputToBucket(str)
	if !t.disableTerminalOutput {
		return PrintlnToTerminal(str)
	}
	return 0, nil
}

func (t *TeePrinter) DisableTerminalOutput(disable bool) {
	t.disableTerminalOutput = disable
}

func (t *TeePrinter) saveOutputToBucket(output string) {
	t.outputBucket.Write([]byte(Decolorize(output)))
}
