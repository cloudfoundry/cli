package translatableerror

type RequiredNameForPushError struct {
}

func (RequiredNameForPushError) DisplayUsage() {}

func (RequiredNameForPushError) Error() string {
	return "Incorrect usage: The push command requires an app name. The app name can be supplied as an argument or with a manifest.yml file."
}

func (e RequiredNameForPushError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
