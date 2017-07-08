package translatableerror

import "fmt"

type NoOrganizationTargetedError struct {
	BinaryName string
}

func (NoOrganizationTargetedError) Error() string {
	return "No org targeted, use '{{.Command}}' to target an org."
}

func (e NoOrganizationTargetedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Command": fmt.Sprintf("%s target -o ORG", e.BinaryName),
	})
}
