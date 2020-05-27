package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

type RequestLoggerFileWriter struct {
	ui            *UI
	lock          *sync.Mutex
	filePaths     []string
	logFiles      []*os.File
	dumpSanitizer *regexp.Regexp
}

func newRequestLoggerFileWriter(ui *UI, lock *sync.Mutex, filePaths []string) *RequestLoggerFileWriter {
	return &RequestLoggerFileWriter{
		ui:            ui,
		lock:          lock,
		filePaths:     filePaths,
		logFiles:      []*os.File{},
		dumpSanitizer: regexp.MustCompile(tokenRegexp),
	}
}

func (display *RequestLoggerFileWriter) DisplayBody([]byte) error {
	for _, logFile := range display.logFiles {
		_, err := logFile.WriteString(RedactedValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func (display *RequestLoggerFileWriter) DisplayDump(dump string) error {
	sanitized := display.dumpSanitizer.ReplaceAllString(dump, RedactedValue)
	cookieCutter := regexp.MustCompile("Set-Cookie:.*")
	sanitized = cookieCutter.ReplaceAllString(sanitized, "Set-Cookie: "+RedactedValue)
	for _, logFile := range display.logFiles {
		_, err := logFile.WriteString(sanitized)
		if err != nil {
			return err
		}
	}
	return nil
}

func (display *RequestLoggerFileWriter) DisplayHeader(name string, value string) error {
	return display.DisplayMessage(fmt.Sprintf("%s: %s", name, value))
}

func (display *RequestLoggerFileWriter) DisplayHost(name string) error {
	return display.DisplayMessage(fmt.Sprintf("Host: %s", name))
}

func (display *RequestLoggerFileWriter) DisplayJSONBody(body []byte) error {
	if len(body) == 0 {
		return nil
	}

	sanitized, err := SanitizeJSON(body)
	if err != nil {
		return display.DisplayMessage(string(body))
	}

	for _, logFile := range display.logFiles {
		_, err = logFile.Write(sanitized)
		if err != nil {
			return err
		}
	}
	return nil
}

func (display *RequestLoggerFileWriter) DisplayMessage(msg string) error {
	for _, logFile := range display.logFiles {
		_, err := logFile.WriteString(fmt.Sprintf("%s\n", msg))
		if err != nil {
			return err
		}
	}
	return nil
}

func (display *RequestLoggerFileWriter) DisplayRequestHeader(method string, uri string, httpProtocol string) error {
	return display.DisplayMessage(fmt.Sprintf("%s %s %s", method, uri, httpProtocol))
}

func (display *RequestLoggerFileWriter) DisplayResponseHeader(httpProtocol string, status string) error {
	return display.DisplayMessage(fmt.Sprintf("%s %s", httpProtocol, status))
}

func (display *RequestLoggerFileWriter) DisplayType(name string, requestDate time.Time) error {
	return display.DisplayMessage(fmt.Sprintf("%s: [%s]", name, requestDate.Format(time.RFC3339)))
}

func (display *RequestLoggerFileWriter) HandleInternalError(err error) {
	display.ui.DisplayWarning(err.Error())
}

func (display *RequestLoggerFileWriter) Start() error {
	display.lock.Lock()
	for _, filePath := range display.filePaths {
		err := os.MkdirAll(filepath.Dir(filePath), os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}

		logFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return err
		}

		display.logFiles = append(display.logFiles, logFile)
	}
	return nil
}

func (display *RequestLoggerFileWriter) Stop() error {
	var err error

	for _, logFile := range display.logFiles {
		_, lastLineErr := logFile.WriteString("\n")
		closeErr := logFile.Close()
		switch {
		case closeErr != nil:
			err = closeErr
		case lastLineErr != nil:
			err = lastLineErr
		}
	}
	display.logFiles = []*os.File{}
	display.lock.Unlock()
	return err
}

// RequestLoggerFileWriter returns a RequestLoggerFileWriter that cannot
// overwrite another RequestLoggerFileWriter.
func (ui *UI) RequestLoggerFileWriter(filePaths []string) *RequestLoggerFileWriter {
	return newRequestLoggerFileWriter(ui, ui.fileLock, filePaths)
}
