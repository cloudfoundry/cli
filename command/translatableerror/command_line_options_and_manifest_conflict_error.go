package translatableerror

import "strings"

type CommandLineOptionsAndManifestConflictError struct {
	ManifestAttribute  string
	CommandLineOptions []string
}

func (e CommandLineOptionsAndManifestConflictError) Error() string {
	return "The following arguments cannot be used with an app manifest that declares routes using the '{{.ManifestAttribute}}' attribute: {{.CommandLineOptions}}"
}

func (e CommandLineOptionsAndManifestConflictError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ManifestAttribute":  e.ManifestAttribute,
		"CommandLineOptions": strings.Join(e.CommandLineOptions, ", "),
	})
}
