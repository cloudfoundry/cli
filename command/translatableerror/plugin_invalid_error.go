package translatableerror

import "fmt"

// PluginInvalidError is returned with a plugin is invalid because it is
// missing a name or has 0 commands.
type PluginInvalidError struct {
	Err error
}

func (e PluginInvalidError) Error() string {
	baseErrString := "File is not a valid cf CLI plugin binary."

	if e.Err != nil {
		return fmt.Sprintf("%s\n%s", e.Err, baseErrString)
	}

	return baseErrString
}

func (e PluginInvalidError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
