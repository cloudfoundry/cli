package actionerror

type CommandLineOptionsWithMultipleAppsError struct{}

func (CommandLineOptionsWithMultipleAppsError) Error() string {
	return "cannot use command line flag with multiple apps"
}
