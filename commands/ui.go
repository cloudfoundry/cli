package commands

// Custom fake was written for this under customv2fakes

// UI is the interface to STDOUT
type UI interface {
	DisplayBoolPrompt(prompt string, defaultResponse bool) (bool, error)
	DisplayError(err error)
	DisplayHeaderFlavorText(text string, keys ...map[string]interface{})
	DisplayHelpHeader(text string)
	DisplayNewline()
	DisplayOK()
	DisplayPair(attribute string, formattedString string, keys ...map[string]interface{})
	DisplayTable(prefix string, table [][]string) error
	DisplayText(template string, data ...map[string]interface{})
	DisplayTextWithKeyTranslations(template string, keysToTranslate []string, data ...map[string]interface{})
	DisplayWarning(formattedString string, keys ...map[string]interface{})
	DisplayWarnings(warnings []string)
}
