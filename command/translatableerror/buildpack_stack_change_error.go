package translatableerror

import "fmt"

type BuildpackStackChangeError struct {
	BuildpackName string
}

func (e BuildpackStackChangeError) Error() string {
	return fmt.Sprintf("Buildpack %s already exists with a stack association", e.BuildpackName)
}

func (e BuildpackStackChangeError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"BuildpackName": e.BuildpackName,
	})
}
