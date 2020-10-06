package lager

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	ERROR
	FATAL
)

var logLevelStr = [...]string{
	DEBUG: "debug",
	INFO:  "info",
	ERROR: "error",
	FATAL: "fatal",
}

func (l LogLevel) String() string {
	if DEBUG <= l && l <= FATAL {
		return logLevelStr[l]
	}
	return "invalid"
}

func LogLevelFromString(s string) (LogLevel, error) {
	for k, v := range logLevelStr {
		if v == s {
			return LogLevel(k), nil
		}
	}
	return -1, fmt.Errorf("invalid log level: %s", s)
}

type Data map[string]interface{}

type rfc3339Time time.Time

const rfc3339Nano = "2006-01-02T15:04:05.000000000Z07:00"

func (t rfc3339Time) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf(`"%s"`, time.Time(t).UTC().Format(rfc3339Nano))
	return []byte(stamp), nil
}

func (t *rfc3339Time) UnmarshalJSON(data []byte) error {
	return (*time.Time)(t).UnmarshalJSON(data)
}

type LogFormat struct {
	Timestamp string   `json:"timestamp"`
	Source    string   `json:"source"`
	Message   string   `json:"message"`
	LogLevel  LogLevel `json:"log_level"`
	Data      Data     `json:"data"`
	Error     error    `json:"-"`
	time      time.Time
}

func (log LogFormat) ToJSON() []byte {
	content, err := json.Marshal(log)
	if err != nil {
		log.Data = dataForJSONMarhallingError(err, log.Data)
		content, err = json.Marshal(log)
		if err != nil {
			panic(err)
		}
	}
	return content
}

func (log LogFormat) toPrettyJSON() []byte {
	t := log.time
	if t.IsZero() {
		t = parseTimestamp(log.Timestamp)
	}

	prettyLog := struct {
		Timestamp rfc3339Time `json:"timestamp"`
		Level     string      `json:"level"`
		Source    string      `json:"source"`
		Message   string      `json:"message"`
		Data      Data        `json:"data"`
		Error     error       `json:"-"`
	}{
		Timestamp: rfc3339Time(t),
		Level:     log.LogLevel.String(),
		Source:    log.Source,
		Message:   log.Message,
		Data:      log.Data,
		Error:     log.Error,
	}

	content, err := json.Marshal(prettyLog)

	if err != nil {
		prettyLog.Data = dataForJSONMarhallingError(err, prettyLog.Data)
		content, err = json.Marshal(prettyLog)
		if err != nil {
			panic(err)
		}
	}

	return content
}

func dataForJSONMarhallingError(err error, data Data) Data {
	_, ok1 := err.(*json.UnsupportedTypeError)
	_, ok2 := err.(*json.MarshalerError)
	errKey := "unknown_error"
	if ok1 || ok2 {
		errKey = "lager serialisation error"
	}

	return map[string]interface{}{
		errKey:      err.Error(),
		"data_dump": fmt.Sprintf("%#v", data),
	}
}

func parseTimestamp(s string) time.Time {
	if s == "" {
		return time.Now()
	}
	n := strings.IndexByte(s, '.')
	if n <= 0 || n == len(s)-1 {
		return time.Now()
	}
	sec, err := strconv.ParseInt(s[:n], 10, 64)
	if err != nil || sec < 0 {
		return time.Now()
	}
	nsec, err := strconv.ParseInt(s[n+1:], 10, 64)
	if err != nil || nsec < 0 {
		return time.Now()
	}
	return time.Unix(sec, nsec)
}
