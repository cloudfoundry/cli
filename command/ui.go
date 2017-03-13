package command

import (
	"time"

	"code.cloudfoundry.org/cli/util/ui"
)

// Custom fake was written for this under customv2fakes

// UI is the interface to STDOUT
type UI interface {
	DisplayBoolPrompt(defaultResponse bool, template string, templateValues ...map[string]interface{}) (bool, error)
	DisplayError(err error)
	DisplayHeader(text string)
	DisplayLogMessage(message ui.LogMessage, displayHeader bool)
	DisplayNewline()
	DisplayOK()
	DisplayTable(prefix string, table [][]string, padding int)
	DisplayText(template string, data ...map[string]interface{})
	DisplayTextWithFlavor(text string, keys ...map[string]interface{})
	DisplayWarning(formattedString string, keys ...map[string]interface{})
	DisplayWarnings(warnings []string)
	RequestLoggerFileWriter(filePaths []string) *ui.RequestLoggerFileWriter
	RequestLoggerTerminalDisplay() *ui.RequestLoggerTerminalDisplay
	TranslateText(template string, data ...map[string]interface{}) string
	UserFriendlyDate(input time.Time) string
}
