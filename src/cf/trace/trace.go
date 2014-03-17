package trace

import (
	"github.com/cloudfoundry/gofileutils/fileutils"
	"io"
	"log"
	"os"
)

const CF_TRACE = "CF_TRACE"

type Printer interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type nullLogger struct{}

func (*nullLogger) Print(v ...interface{})                 {}
func (*nullLogger) Printf(format string, v ...interface{}) {}
func (*nullLogger) Println(v ...interface{})               {}

var stdOut io.Writer = os.Stdout
var Logger Printer

func init() {
	Logger = NewLogger()
}

func EnableTrace() {
	Logger = newStdoutLogger()
}

func DisableTrace() {
	Logger = new(nullLogger)
}

func SetStdout(s io.Writer) {
	stdOut = s
}

func NewLogger() Printer {
	cf_trace := os.Getenv(CF_TRACE)
	switch cf_trace {
	case "", "false":
		return new(nullLogger)
	case "true":
		return newStdoutLogger()
	default:
		return newFileLogger(cf_trace)
	}
}

func newStdoutLogger() Printer {
	return log.New(stdOut, "", 0)
}

func newFileLogger(path string) Printer {
	file, err := fileutils.Open(path)
	if err != nil {
		logger := newStdoutLogger()
		logger.Printf("CF_TRACE ERROR CREATING LOG FILE %s:\n%s", path, err)
		return logger
	}

	return log.New(file, "", 0)
}
