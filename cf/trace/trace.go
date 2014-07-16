package trace

import (
	. "github.com/cloudfoundry/cli/cf/i18n"
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
	Logger = NewLogger("")
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

func NewLogger(cf_trace string) Printer {
	switch cf_trace {
	case "", "false":
		Logger = new(nullLogger)
	case "true":
		Logger = newStdoutLogger()
	default:
		Logger = newFileLogger(cf_trace)
	}

	return Logger
}

func newStdoutLogger() Printer {
	return log.New(stdOut, "", 0)
}

func newFileLogger(path string) Printer {
	file, err := fileutils.Open(path)
	if err != nil {
		logger := newStdoutLogger()
		logger.Printf(T("CF_TRACE ERROR CREATING LOG FILE {{.Path}}:\n{{.Err}}",
			map[string]interface{}{"Path": path, "Err": err}))
		return logger
	}

	return log.New(file, "", 0)
}
