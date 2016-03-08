package trace

import (
	"io"
	"log"
)

func NewWriterPrinter(writer io.Writer) Printer {
	return log.New(writer, "", 0)
}
