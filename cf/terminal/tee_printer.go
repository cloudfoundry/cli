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

func (t *TeePrinter) Print(values ...interface{}) (n int, err error) {
	str := fmt.Sprint(values...)
	t.saveOutputToBucket(str)
	if !t.disableTerminalOutput {
		return fmt.Print(str)
	}
	return
}

func (t *TeePrinter) Printf(format string, a ...interface{}) (n int, err error) {
	str := fmt.Sprintf(format, a...)
	t.saveOutputToBucket(str)
	if !t.disableTerminalOutput {
		return fmt.Print(str)
	}
	return
}

func (t *TeePrinter) Println(values ...interface{}) (n int, err error) {
	str := fmt.Sprint(values...)
	t.saveOutputToBucket(str)
	if !t.disableTerminalOutput {
		return fmt.Println(str)
	}
	return
}

func (t *TeePrinter) DisableTerminalOutput(disable bool) {
	t.disableTerminalOutput = disable
}

func (t *TeePrinter) saveOutputToBucket(output string) {
	t.outputBucket.Write([]byte(Decolorize(output)))
}
