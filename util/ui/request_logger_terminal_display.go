package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/fatih/color"
)

type RequestLoggerTerminalDisplay struct {
	ui            *UI
	lock          *sync.Mutex
	dumpSanitizer *regexp.Regexp
}

func newRequestLoggerTerminalDisplay(ui *UI, lock *sync.Mutex) *RequestLoggerTerminalDisplay {
	return &RequestLoggerTerminalDisplay{
		ui:            ui,
		lock:          lock,
		dumpSanitizer: regexp.MustCompile(tokenRegexp),
	}
}

func (display *RequestLoggerTerminalDisplay) DisplayBody([]byte) error {
	fmt.Fprintf(display.ui.Out, "%s\n", RedactedValue)
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayDump(dump string) error {
	sanitized := display.dumpSanitizer.ReplaceAllString(dump, RedactedValue)
	fmt.Fprintf(display.ui.Out, "%s\n", sanitized)
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayHeader(name string, value string) error {
	fmt.Fprintf(display.ui.Out, "%s: %s\n", display.ui.TranslateText(name), value)
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayHost(name string) error {
	fmt.Fprintf(display.ui.Out, "%s: %s\n", display.ui.TranslateText("Host"), name)
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayJSONBody(body []byte) error {
	if body == nil || len(body) == 0 {
		return nil
	}

	sanitized, err := SanitizeJSON(body)
	if err != nil {
		fmt.Fprintf(display.ui.Out, "%s\n", string(body))
		return nil
	}

	buff := new(bytes.Buffer)
	encoder := json.NewEncoder(buff)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(sanitized)
	if err != nil {
		fmt.Fprintf(display.ui.Out, "%s\n", string(body))
	}

	fmt.Fprintf(display.ui.Out, "%s\n", buff.String())
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayMessage(msg string) error {
	fmt.Fprintf(display.ui.Out, "%s\n", msg)
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayRequestHeader(method string, uri string, httpProtocol string) error {
	fmt.Fprintf(display.ui.Out, "%s %s %s\n", method, uri, httpProtocol)
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayResponseHeader(httpProtocol string, status string) error {
	fmt.Fprintf(display.ui.Out, "%s %s\n", httpProtocol, status)
	return nil
}

func (display *RequestLoggerTerminalDisplay) DisplayType(name string, requestDate time.Time) error {
	text := fmt.Sprintf("%s: [%s]", name, requestDate.Format(time.RFC3339))
	fmt.Fprintf(display.ui.Out, "%s\n", display.ui.modifyColor(display.ui.TranslateText(text), color.New(color.Bold)))
	return nil
}

func (display *RequestLoggerTerminalDisplay) HandleInternalError(err error) {
	fmt.Fprintf(display.ui.Err, "%s\n", display.ui.TranslateText(err.Error()))
}

func (display *RequestLoggerTerminalDisplay) Start() error {
	display.lock.Lock()
	return nil
}

func (display *RequestLoggerTerminalDisplay) Stop() error {
	fmt.Fprintf(display.ui.Out, "\n")
	display.lock.Unlock()
	return nil
}
