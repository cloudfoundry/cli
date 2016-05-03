package gosteno

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	EXCLUDE_NONE = 0

	EXCLUDE_LEVEL = 1 << (iota - 1)
	EXCLUDE_TIMESTAMP
	EXCLUDE_FILE
	EXCLUDE_LINE
	EXCLUDE_METHOD
	EXCLUDE_DATA
	EXCLUDE_MESSAGE
)

type JsonPrettifier struct {
	entryTemplate *template.Template
}

func NewJsonPrettifier(flag int) *JsonPrettifier {
	fields := []string{
		"{{encodeLevel .Level}}",
		"{{encodeTimestamp .Timestamp}}",
		"{{encodeFile .File}}",
		"{{encodeLine .Line}}",
		"{{encodeMethod .Method}}",
		"{{encodeData .Data}}",
		"{{encodeMessage .Message}}",
	}

	for i, _ := range fields {
		// the shift count must be an unsigned integer
		if (flag & (1 << uint(i))) != 0 {
			fields[i] = ""
		}
	}

	prettifier := new(JsonPrettifier)
	format := strings.Join(fields, "")
	funcMap := template.FuncMap{
		"encodeTimestamp": encodeTimestamp,
		"encodeFile":      encodeFile,
		"encodeMethod":    encodeMethod,
		"encodeLine":      encodeLine,
		"encodeData":      encodeData,
		"encodeLevel":     encodeLevel,
		"encodeMessage":   encodeMessage,
	}
	prettifier.entryTemplate = template.Must(template.New("EntryTemplate").Funcs(funcMap).Parse(format))

	return prettifier
}

func (p *JsonPrettifier) DecodeJsonLogEntry(logEntry string) (*Record, error) {
	record := new(Record)
	err := json.Unmarshal([]byte(logEntry), record)
	return record, err
}

func (p *JsonPrettifier) EncodeRecord(record *Record) ([]byte, error) {
	buffer := bytes.NewBufferString("")
	err := p.entryTemplate.Execute(buffer, record)
	return buffer.Bytes(), err
}

func encodeLevel(level LogLevel) string {
	return fmt.Sprintf("%s ", strings.ToUpper(level.String()))
}

func encodeTimestamp(t RecordTimestamp) string {
	ut := time.Unix(int64(t), 0)
	return fmt.Sprintf("%s ", ut.Format("2006-01-02 15:04:05"))
}

func encodeFile(file string) string {
	index := strings.LastIndex(file, "/")
	return fmt.Sprintf("%s:", file[index+1:])
}

func encodeLine(line int) string {
	return fmt.Sprintf("%s:", strconv.Itoa(line))
}

func encodeMethod(method string) string {
	index := strings.LastIndex(method, ".")
	return fmt.Sprintf("%s ", method[index+1:])
}

func encodeData(data map[string]interface{}) (string, error) {
	b, err := json.Marshal(data)
	return fmt.Sprintf("%s ", string(b)), err
}

func encodeMessage(message string) string {
	return message
}
