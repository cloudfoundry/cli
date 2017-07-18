package translatableerror

type CommandLineArgsWithMultipleAppsError struct {
}

func (CommandLineArgsWithMultipleAppsError) Error() string {
	return "Incorrect Usage: Command line flags (except -f) cannot be applied when pushing multiple apps from a manifest file."
}

func (e CommandLineArgsWithMultipleAppsError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
