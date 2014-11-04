package terminal

import (
	"fmt"
)

type Printer interface {
	Print(a ...interface{}) (n int, err error)
	Printf(format string, a ...interface{}) (n int, err error)
	Println(a ...interface{}) (n int, err error)
}

type OutputCapture interface {
	GetOutputAndReset() []string
}

type TerminalOutputSwitch interface {
	PauseTerminalOutput()
	ResumeTerminalOutput()
}

type TeePrinter struct {
	pause  bool
	output []string
}

func NewTeePrinter() *TeePrinter {
	return &TeePrinter{
		output: []string{},
	}
}

func (t *TeePrinter) GetOutputAndReset() []string {
	currentOutput := t.output
	t.output = []string{}
	return currentOutput
}

func (t *TeePrinter) Print(values ...interface{}) (n int, err error) {
	str := fmt.Sprint(values...)
	t.output = append(t.output, str)
	if !t.pause {
		return fmt.Print(str)
	}
	return
}

func (t *TeePrinter) Printf(format string, a ...interface{}) (n int, err error) {
	str := fmt.Sprintf(format, a...)
	t.output = append(t.output, str)
	if !t.pause {
		return fmt.Print(str)
	}
	return
}

func (t *TeePrinter) Println(values ...interface{}) (n int, err error) {
	str := fmt.Sprint(values...)
	t.output = append(t.output, str)
	if !t.pause {
		return fmt.Println(str)
	}
	return
}

func (t *TeePrinter) PauseTerminalOutput() {
	t.pause = true
}

func (t *TeePrinter) ResumeTerminalOutput() {
	t.pause = false
}
