package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type RequestLoggerFileWriter struct {
	ui        UI
	filePaths []string
	logFiles  []*os.File
}

func NewRequestLoggerFileWriter(ui UI, filePaths []string) *RequestLoggerFileWriter {
	return &RequestLoggerFileWriter{
		ui:        ui,
		filePaths: filePaths,
		logFiles:  []*os.File{},
	}
}

func (display *RequestLoggerFileWriter) DisplayBody(_ []byte) error {
	for _, logFile := range display.logFiles {
		_, err := logFile.WriteString(RedactedValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func (display *RequestLoggerFileWriter) DisplayJSONBody(body []byte) error {
	if body == nil || len(body) == 0 {
		return nil
	}

	sanitized, err := SanitizeJSON(body)
	if err != nil {
		return err
	}

	buff := new(bytes.Buffer)
	encoder := json.NewEncoder(buff)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(sanitized)
	if err != nil {
		return err
	}

	for _, logFile := range display.logFiles {
		_, err = logFile.Write(buff.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (display *RequestLoggerFileWriter) DisplayHeader(name string, value string) error {
	for _, logFile := range display.logFiles {
		_, err := logFile.WriteString(fmt.Sprintf("%s: %s\n", name, value))
		if err != nil {
			return err
		}
	}
	return nil
}

func (display *RequestLoggerFileWriter) DisplayHost(name string) error {
	for _, logFile := range display.logFiles {
		_, err := logFile.WriteString(fmt.Sprintf("Host: %s\n", name))
		if err != nil {
			return err
		}
	}
	return nil
}

func (display *RequestLoggerFileWriter) DisplayRequestHeader(method string, uri string, httpProtocol string) error {
	for _, logFile := range display.logFiles {
		_, err := logFile.WriteString(fmt.Sprintf("%s %s %s\n", method, uri, httpProtocol))
		if err != nil {
			return err
		}
	}
	return nil
}

func (display *RequestLoggerFileWriter) DisplayResponseHeader(httpProtocol string, status string) error {
	for _, logFile := range display.logFiles {
		_, err := logFile.WriteString(fmt.Sprintf("%s %s\n", httpProtocol, status))
		if err != nil {
			return err
		}
	}
	return nil
}

func (display *RequestLoggerFileWriter) DisplayType(name string, requestDate time.Time) error {
	for _, logFile := range display.logFiles {
		_, err := logFile.WriteString(fmt.Sprintf("%s: [%s]\n", name, requestDate.Format(time.RFC3339)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (display *RequestLoggerFileWriter) HandleInternalError(err error) {
	display.ui.DisplayWarning(err.Error())
}

func (display *RequestLoggerFileWriter) Start() error {
	for _, filePath := range display.filePaths {
		logFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return err
		}

		display.logFiles = append(display.logFiles, logFile)
	}
	return nil
}

func (display *RequestLoggerFileWriter) Stop() error {
	for _, logFile := range display.logFiles {
		logFile.WriteString("\n")
		err := logFile.Close()
		if err != nil {
			return err
		}
	}
	display.logFiles = []*os.File{}
	return nil
}
