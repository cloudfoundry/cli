package common

import (
	"encoding/json"
	"fmt"
	"time"
)

type RequestLoggerTerminalDisplay struct {
	ui TerminalDisplay
}

func NewRequestLoggerTerminalDisplay(ui TerminalDisplay) *RequestLoggerTerminalDisplay {
	return &RequestLoggerTerminalDisplay{
		ui: ui,
	}
}

func (display *RequestLoggerTerminalDisplay) DisplayBody(body []byte) {
	sanitized, err := SanitizeJSON(body)
	if err != nil {
		display.ui.DisplayText(string(body))
	}

	pretty, err := json.MarshalIndent(sanitized, "", "  ")
	if err != nil {
		display.ui.DisplayText(string(body))
	}

	display.ui.DisplayText(string(pretty))
}

func (display *RequestLoggerTerminalDisplay) DisplayHeader(name string, value string) {
	display.ui.DisplayPair(name, value)
}

func (display *RequestLoggerTerminalDisplay) DisplayHost(name string) {
	display.ui.DisplayPair("Host", name)
}

func (display *RequestLoggerTerminalDisplay) DisplayRequestHeader(method string, uri string, httpProtocol string) {
	display.ui.DisplayText("{{.Method}} {{.URI}} {{.Proto}}}", map[string]interface{}{
		"Method": method,
		"URI":    uri,
		"Proto":  httpProtocol,
	})
}

func (display *RequestLoggerTerminalDisplay) DisplayResponseHeader(httpProtocol string, status string) {
	display.ui.DisplayText("{{.Proto}} {{.Status}}", map[string]interface{}{
		"Proto":  httpProtocol,
		"Status": status,
	})
}

func (display *RequestLoggerTerminalDisplay) DisplayType(name string, requestDate time.Time) {
	display.ui.DisplayPair(name, fmt.Sprintf("[%s]", requestDate.Format(time.RFC3339)))
}

func (display *RequestLoggerTerminalDisplay) HandleInternalError(err error) {
	display.ui.DisplayErrorMessage(err.Error())
}

func (display *RequestLoggerTerminalDisplay) Start() error { return nil }

func (display *RequestLoggerTerminalDisplay) Stop() error {
	display.ui.DisplayNewline()
	return nil
}
