package translatableerror

// ArgumentManifestMismatchError represent an error caused by using a command line flag
// that conflicts with a given manifest property.
type ArgumentManifestMismatchError struct {
	Arg              string
	ManifestProperty string
	ManifestValue    string
}

func (ArgumentManifestMismatchError) DisplayUsage() {}

func (e ArgumentManifestMismatchError) Error() string {
	if e.ManifestValue == "" {
		return "Incorrect Usage: The flag option {{.Arg}} cannot be used with the manifest property {{.Property}}"
	}
	return "Incorrect Usage: The flag option {{.Arg}} cannot be used with the manifest property {{.Property}} set to {{.ManifestValue}}"
}

func (e ArgumentManifestMismatchError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Arg":           e.Arg,
		"Property":      e.ManifestProperty,
		"ManifestValue": e.ManifestValue,
	})
}
