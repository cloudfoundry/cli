package gosteno

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

type RecordTimestamp float64

func (t RecordTimestamp) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.9f", t)), nil
}

type Record struct {
	Timestamp RecordTimestamp        `json:"timestamp"`
	Pid       int                    `json:"process_id"`
	Source    string                 `json:"source"`
	Level     LogLevel               `json:"log_level"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
	Method    string                 `json:"method,omitempty"`
}

var pid int

func init() {
	pid = os.Getpid()
}

func NewRecord(s string, l LogLevel, m string, d map[string]interface{}) *Record {
	r := &Record{
		Timestamp: RecordTimestamp(time.Now().UnixNano()) / 1000000000,
		Pid:       pid,
		Source:    s,
		Level:     l,
		Message:   m,
		Data:      d,
	}

	if getConfig().EnableLOC {
		var function *runtime.Func
		var file string
		var line int

		pc := make([]uintptr, 50)
		nptrs := runtime.Callers(2, pc)
		for i := 0; i < nptrs; i++ {
			function = runtime.FuncForPC(pc[i])
			file, line = function.FileLine(pc[i])
			if !strings.HasSuffix(file, "logger.go") {
				break
			}
		}
		r.File = file
		r.Line = line
		r.Method = function.Name()
	}

	return r
}
