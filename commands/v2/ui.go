package v2

// Custom fake was written for this under customv2fakes

// UI is the interface to STDOUT
type UI interface {
	DisplayText(string, ...map[string]interface{})
	DisplayTextWithKeyTranslations(string, []string, ...map[string]interface{})
	DisplayNewline()
	DisplayHelpHeader(string)
}
