package translatableerror

import "fmt"

type NoAPISetError struct {
	BinaryName string
}

func (NoAPISetError) Error() string {
	return "No API endpoint set. Use '{{.LoginTip}}' or '{{.APITip}}' to target an endpoint."
}

func (e NoAPISetError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"LoginTip": fmt.Sprintf("%s login", e.BinaryName),
		"APITip":   fmt.Sprintf("%s api", e.BinaryName),
	})
}
