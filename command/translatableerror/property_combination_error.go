package translatableerror

import (
	"strings"
)

type PropertyCombinationError struct {
	AppName    string
	Properties []string
}

func (e PropertyCombinationError) Error() string {
	return "Application {{.AppName}} cannot use the combination of properties: {{.Properties}}"
}

func (e PropertyCombinationError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName":    e.AppName,
		"Properties": strings.Join(e.Properties, ", "),
	})
}
