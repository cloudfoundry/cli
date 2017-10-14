package command

import (
	"io"
	"time"

	"code.cloudfoundry.org/cli/util/ui"
)

// UI is the interface to STDOUT
type UI interface {
	DisplayBoolPrompt(defaultResponse bool, template string, templateValues ...map[string]interface{}) (bool, error)
	DisplayPasswordPrompt(template string, templateValues ...map[string]interface{}) (string, error)
	DisplayChangesForPush(changeSet []ui.Change) error
	DisplayError(err error)
	DisplayHeader(text string)
	DisplayInstancesTableForApp(table [][]string)
	DisplayKeyValueTable(prefix string, table [][]string, padding int)
	DisplayKeyValueTableForApp(table [][]string)
	DisplayKeyValueTableForV3App(table [][]string, crashedProcesses []string)
	DisplayLogMessage(message ui.LogMessage, displayHeader bool)
	DisplayNewline()
	DisplayNonWrappingTable(prefix string, table [][]string, padding int)
	DisplayOK()
	DisplayTableWithHeader(prefix string, table [][]string, padding int)
	DisplayText(template string, data ...map[string]interface{})
	DisplayTextWithFlavor(text string, keys ...map[string]interface{})
	DisplayTextWithBold(text string, keys ...map[string]interface{})
	DisplayWarning(formattedString string, keys ...map[string]interface{})
	DisplayWarnings(warnings []string)
	GetIn() io.Reader
	GetOut() io.Writer
	GetErr() io.Writer
	RequestLoggerFileWriter(filePaths []string) *ui.RequestLoggerFileWriter
	RequestLoggerTerminalDisplay() *ui.RequestLoggerTerminalDisplay
	TranslateText(template string, data ...map[string]interface{}) string
	UserFriendlyDate(input time.Time) string
	Writer() io.Writer
}
