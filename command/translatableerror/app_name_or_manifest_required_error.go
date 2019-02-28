package translatableerror

// ArgumentCombinationError represent an error caused by using two command line
// arguments that cannot be used together.
type AppNameOrManifestRequiredError struct {
	Args []string
}

func (AppNameOrManifestRequiredError) DisplayUsage() {}

func (AppNameOrManifestRequiredError) Error() string {
	return "Incorrect Usage:  The push command requires an app name. The app name can be supplied as an argument or with a manifest.yml file."
}

func (e AppNameOrManifestRequiredError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
