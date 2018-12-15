package translatableerror

import "fmt"

type BuildpackStackChangeError struct {
	BuildpackName string
	BinaryName    string
}

func (e BuildpackStackChangeError) Error() string {
	return fmt.Sprintf("Buildpack {{.BuildpackName}} already exists with a stack association\n\nTIP: Use '{{.BuildpackCommand}}' to view buildpack and stack associations")
}

func (e BuildpackStackChangeError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"BuildpackName":    e.BuildpackName,
		"BuildpackCommand": fmt.Sprintf("%s buildpacks", e.BinaryName),
	})
}
