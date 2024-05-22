package command

import (
	"io"
	"time"

	"code.cloudfoundry.org/cli/util/ui"
)

// UI is the interface to STDOUT, STDERR, and STDIN.
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . UI
type UI interface {
	DeferText(template string, data ...map[string]interface{})
	DisplayBoolPrompt(defaultResponse bool, template string, templateValues ...map[string]interface{}) (bool, error)
	DisplayChangesForPush(changeSet []ui.Change) error
	DisplayDeprecationWarning()
	DisplayDiffAddition(line string, depth int, addHyphen bool)
	DisplayDiffRemoval(line string, depth int, addHyphen bool)
	DisplayDiffUnchanged(line string, depth int, addHyphen bool)
	DisplayError(err error)
	DisplayFileDeprecationWarning()
	DisplayHeader(text string)
	DisplayInstancesTableForApp(table [][]string)
	DisplayJSON(name string, jsonData interface{}) error
	DisplayKeyValueTable(prefix string, table [][]string, padding int)
	DisplayKeyValueTableForApp(table [][]string)
	DisplayLogMessage(message ui.LogMessage, displayHeader bool)
	DisplayNewline()
	DisplayNonWrappingTable(prefix string, table [][]string, padding int)
	DisplayOK()
	DisplayOptionalTextPrompt(defaultValue string, template string, templateValues ...map[string]interface{}) (string, error)
	DisplayPasswordPrompt(template string, templateValues ...map[string]interface{}) (string, error)
	DisplayTableWithHeader(prefix string, table [][]string, padding int)
	DisplayText(template string, data ...map[string]interface{})
	DisplayTextMenu(choices []string, promptTemplate string, templateValues ...map[string]interface{}) (string, error)
	DisplayTextPrompt(template string, templateValues ...map[string]interface{}) (string, error)
	DisplayTextWithBold(text string, keys ...map[string]interface{})
	DisplayTextWithFlavor(text string, keys ...map[string]interface{})
	DisplayWarning(formattedString string, keys ...map[string]interface{})
	DisplayWarnings(warnings []string)
	GetErr() io.Writer
	GetIn() io.Reader
	GetOut() io.Writer
	RequestLoggerFileWriter(filePaths []string) *ui.RequestLoggerFileWriter
	RequestLoggerTerminalDisplay() *ui.RequestLoggerTerminalDisplay
	TranslateText(template string, data ...map[string]interface{}) string
	UserFriendlyDate(input time.Time) string
	Writer() io.Writer
}
