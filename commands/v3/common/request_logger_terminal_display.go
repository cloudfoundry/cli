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

func (display *RequestLoggerTerminalDisplay) DisplayBody(body []byte) error {
	sanitized, err := SanitizeJSON(body)
	if err != nil {
		display.ui.DisplayText(string(body))
	}

	pretty, err := json.MarshalIndent(sanitized, "", "  ")
	if err != nil {
		display.ui.DisplayText(string(body))
	}

	display.ui.DisplayText(string(pretty))
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayHeader(name string, value string) error {
	display.ui.DisplayPair(name, value)
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayHost(name string) error {
	display.ui.DisplayPair("Host", name)
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayRequestHeader(method string, uri string, httpProtocol string) error {
	display.ui.DisplayText("{{.Method}} {{.URI}} {{.Proto}}}", map[string]interface{}{
		"Method": method,
		"URI":    uri,
		"Proto":  httpProtocol,
	})
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayResponseHeader(httpProtocol string, status string) error {
	display.ui.DisplayText("{{.Proto}} {{.Status}}", map[string]interface{}{
		"Proto":  httpProtocol,
		"Status": status,
	})
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayType(name string, requestDate time.Time) error {
	display.ui.DisplayPair(name, fmt.Sprintf("[%s]", requestDate.Format(time.RFC3339)))
	return nil
}

func (display *RequestLoggerTerminalDisplay) HandleInternalError(err error) {
	display.ui.DisplayError(err)
}

func (display *RequestLoggerTerminalDisplay) Start() error { return nil }

func (display *RequestLoggerTerminalDisplay) Stop() error {
	display.ui.DisplayNewline()
	return nil
}
