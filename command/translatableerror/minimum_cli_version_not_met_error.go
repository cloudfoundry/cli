package translatableerror

type MinimumCLIVersionNotMetError struct {
	APIVersion    string
	MinCLIVersion string
	BinaryVersion string
}

func (e MinimumCLIVersionNotMetError) Error() string {
	return "Cloud Foundry API version {{.APIVersion}} requires CLI version {{.MinCLIVersion}}. You are currently on version {{.BinaryVersion}}. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads"
}

func (e MinimumCLIVersionNotMetError) Translate(translate func(string, ...interface{}) string) string {
	vars := map[string]interface{}{
		"APIVersion":    e.APIVersion,
		"MinCLIVersion": e.MinCLIVersion,
		"BinaryVersion": e.BinaryVersion,
	}

	return translate(e.Error(), vars)
}
