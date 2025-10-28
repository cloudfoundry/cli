package translatableerror

import "fmt"

type BuildpackNotFoundError struct {
	BuildpackName string
	StackName     string
	Lifecycle     string
}

func (e BuildpackNotFoundError) Error() string {
	message := fmt.Sprintf("Buildpack '%s'", e.BuildpackName)
	if len(e.StackName) != 0 {
		message = fmt.Sprintf("%s with stack '%s'", message, e.StackName)
	}
	if len(e.Lifecycle) != 0 {
		message = fmt.Sprintf("%s with lifecycle '%s'", message, e.Lifecycle)
	}
	return message + " not found"
}

func (e BuildpackNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"BuildpackName": e.BuildpackName,
		"StackName":     e.StackName,
		"Lifecycle":     e.Lifecycle,
	})
}
