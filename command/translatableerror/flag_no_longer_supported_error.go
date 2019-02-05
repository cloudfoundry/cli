package translatableerror

// FlagNoLongerSupportedError can be used to indicate a flag is no longer supported
type FlagNoLongerSupportedError struct {
	Flag string
}

func (e FlagNoLongerSupportedError) Error() string {
	return "Flag '{{.Flag}}' is no longer supported."
}

func (e FlagNoLongerSupportedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Flag": e.Flag,
	})
}
