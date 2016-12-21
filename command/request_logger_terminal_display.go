package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type RequestLoggerTerminalDisplay struct {
	ui UI
}

func NewRequestLoggerTerminalDisplay(ui UI) *RequestLoggerTerminalDisplay {
	return &RequestLoggerTerminalDisplay{
		ui: ui,
	}
}

func (display *RequestLoggerTerminalDisplay) DisplayBody(_ []byte) error {
	display.ui.DisplayText(RedactedValue)
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayJSONBody(body []byte) error {
	if body == nil || len(body) == 0 {
		return nil
	}

	sanitized, err := SanitizeJSON(body)
	if err != nil {
		display.ui.DisplayText(string(body))
	}

	buff := new(bytes.Buffer)
	encoder := json.NewEncoder(buff)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(sanitized)
	if err != nil {
		display.ui.DisplayText(string(body))
	}

	display.ui.DisplayText(buff.String())
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
	display.ui.DisplayText("{{.Method}} {{.URI}} {{.Proto}}", map[string]interface{}{
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
	display.ui.DisplayHeader(fmt.Sprintf("%s: [%s]", name, requestDate.Format(time.RFC3339)))
	return nil
}

func (display *RequestLoggerTerminalDisplay) HandleInternalError(err error) {
	display.ui.DisplayWarning(err.Error())
}

func (display *RequestLoggerTerminalDisplay) Start() error { return nil }

func (display *RequestLoggerTerminalDisplay) Stop() error {
	display.ui.DisplayNewline()
	return nil
}
