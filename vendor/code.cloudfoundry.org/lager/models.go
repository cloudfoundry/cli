package lager

import (
	"encoding/json"
	"fmt"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	ERROR
	FATAL
)

type Data map[string]interface{}

type LogFormat struct {
	Timestamp string   `json:"timestamp"`
	Source    string   `json:"source"`
	Message   string   `json:"message"`
	LogLevel  LogLevel `json:"log_level"`
	Data      Data     `json:"data"`
}

func (log LogFormat) ToJSON() []byte {
	content, err := json.Marshal(log)
	if err != nil {
		if _, ok := err.(*json.UnsupportedTypeError); ok {
			log.Data = map[string]interface{}{"lager serialisation error": err.Error(), "data_dump": fmt.Sprintf("%#v", log.Data)}
			content, err = json.Marshal(log)
		}
		if err != nil {
			panic(err)
		}
	}
	return content
}
