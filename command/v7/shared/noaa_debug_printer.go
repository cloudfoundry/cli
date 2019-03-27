package shared

import (
	"time"
)

type RequestLoggerOutput interface {
	Start() error
	Stop() error
	DisplayType(name string, requestDate time.Time) error
	DisplayDump(dump string) error
}

type NOAADebugPrinter struct {
	outputs []RequestLoggerOutput
}

func (p *NOAADebugPrinter) addOutput(output RequestLoggerOutput) {
	p.outputs = append(p.outputs, output)
}

func (p NOAADebugPrinter) Print(title string, dump string) {
	for _, output := range p.outputs {
		_ = output.Start()
		defer output.Stop()

		output.DisplayType(title, time.Now())
		output.DisplayDump(dump)
	}
}
