package common

// Use custom UI fake instead of counterfeiter fake

type TerminalDisplay interface {
	DisplayError(err error)
	DisplayNewline()
	DisplayPair(attribute string, formattedString string, keys ...map[string]interface{})
	DisplayText(template string, data ...map[string]interface{})
}
