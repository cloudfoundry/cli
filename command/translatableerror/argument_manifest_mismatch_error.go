package translatableerror

// ArgumentManifestMismatchError represent an error caused by using a command line flag
// that conflicts with a given manifest property.
type ArgumentManifestMismatchError struct {
	Arg              string
	ManifestProperty string
}

func (ArgumentManifestMismatchError) DisplayUsage() {}

func (ArgumentManifestMismatchError) Error() string {
	return "Incorrect Usage: The argument {{.Arg}} cannot be used with the manifest property {{.Property}}"
}

func (e ArgumentManifestMismatchError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Arg": e.Arg,
		"Property": e.ManifestProperty,
	})
}
