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

type TeePrinter struct {
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
	return fmt.Print(str)
}

func (t *TeePrinter) Printf(format string, a ...interface{}) (n int, err error) {
	str := fmt.Sprintf(format, a...)
	t.output = append(t.output, str)
	return fmt.Print(str)
}

func (t *TeePrinter) Println(values ...interface{}) (n int, err error) {
	str := fmt.Sprint(values...)
	t.output = append(t.output, str)
	return fmt.Println(str)
}
