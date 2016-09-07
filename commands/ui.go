package commands

// Custom fake was written for this under customv2fakes

// UI is the interface to STDOUT
type UI interface {
	DisplayText(template string, data ...map[string]interface{})
	DisplayTextWithKeyTranslations(template string, keysToTranslate []string, data ...map[string]interface{})
	DisplayNewline()
	DisplayHelpHeader(text string)
	DisplayTable(prefix string, table [][]string)
	DisplayHeaderFlavorText(text string, keys ...map[string]interface{})
}
