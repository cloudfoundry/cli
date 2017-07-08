package translatableerror

import "fmt"

type NoSpaceTargetedError struct {
	BinaryName string
}

func (NoSpaceTargetedError) Error() string {
	return "No space targeted, use '{{.Command}}' to target a space."
}

func (e NoSpaceTargetedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Command": fmt.Sprintf("%s target -s SPACE", e.BinaryName),
	})
}
