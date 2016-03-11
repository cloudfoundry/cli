package trace

import (
	"io"
	"log"
)

type LoggerPrinter struct {
	logger *log.Logger
}

func NewWriterPrinter(writer io.Writer) Printer {
	return LoggerPrinter{
		logger: log.New(writer, "", 0),
	}
}

func (p LoggerPrinter) Print(v ...interface{}) {
	p.logger.Print(v...)
}

func (p LoggerPrinter) Printf(format string, v ...interface{}) {
	p.logger.Printf(format, v...)
}

func (p LoggerPrinter) Println(v ...interface{}) {
	p.logger.Println(v...)
}

func (p LoggerPrinter) IsEnabled() bool {
	return true
}
