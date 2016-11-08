package common

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type RequestLoggerFileWriter struct {
	ui       TerminalDisplay
	filePath string
	logFile  *os.File
}

func NewRequestLoggerFileWriter(ui TerminalDisplay, filePath string) *RequestLoggerFileWriter {
	return &RequestLoggerFileWriter{
		ui:       ui,
		filePath: filePath,
	}
}

func (display *RequestLoggerFileWriter) DisplayBody(body []byte) error {
	sanitized, err := SanitizeJSON(body)
	if err != nil {
		return err
	}

	pretty, err := json.MarshalIndent(sanitized, "", "  ")
	if err != nil {
		return err
	}

	_, err = display.logFile.WriteString(string(pretty) + "\n")
	return err
}

func (display *RequestLoggerFileWriter) DisplayHeader(name string, value string) error {
	_, err := display.logFile.WriteString(fmt.Sprintf("%s: %s\n", name, value))
	return err
}

func (display *RequestLoggerFileWriter) DisplayHost(name string) error {
	_, err := display.logFile.WriteString(fmt.Sprintf("Host: %s\n", name))
	return err
}

func (display *RequestLoggerFileWriter) DisplayRequestHeader(method string, uri string, httpProtocol string) error {
	_, err := display.logFile.WriteString(fmt.Sprintf("%s %s %s\n", method, uri, httpProtocol))
	return err
}

func (display *RequestLoggerFileWriter) DisplayResponseHeader(httpProtocol string, status string) error {
	_, err := display.logFile.WriteString(fmt.Sprintf("%s %s\n", httpProtocol, status))
	return err
}

func (display *RequestLoggerFileWriter) DisplayType(name string, requestDate time.Time) error {
	_, err := display.logFile.WriteString(fmt.Sprintf("%s: [%s]\n", name, requestDate.Format(time.RFC3339)))
	return err
}

func (display *RequestLoggerFileWriter) HandleInternalError(err error) {
	display.ui.DisplayError(err)
}

func (display *RequestLoggerFileWriter) Start() error {
	logFile, err := os.OpenFile(display.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	display.logFile = logFile
	return nil
}

func (display *RequestLoggerFileWriter) Stop() error {
	display.logFile.WriteString("\n")
	return display.logFile.Close()
}
