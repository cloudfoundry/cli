package translatableerror

import "fmt"

type StagingFailedNoAppDetectedError struct {
	Message    string
	BinaryName string
}

func (StagingFailedNoAppDetectedError) Error() string {
	return "Error staging application: {{.Message}}\n\nTIP: Use '{{.BuildpackCommand}}' to see a list of supported buildpacks."
}

func (e StagingFailedNoAppDetectedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Message":          e.Message,
		"BuildpackCommand": fmt.Sprintf("%s buildpacks", e.BinaryName),
	})
}
