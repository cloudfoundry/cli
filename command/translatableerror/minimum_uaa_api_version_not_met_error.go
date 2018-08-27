package translatableerror

type MinimumUAAAPIVersionNotMetError struct {
	Command        string
	MinimumVersion string
}

func (e MinimumUAAAPIVersionNotMetError) Error() string {
	if e.Command != "" {
		return "{{.Command}} requires UAA API version {{.MinimumVersion}} or higher. Update your Cloud Foundry instance."
	}

	return "This command requires UAA API version {{.MinimumVersion}} or higher. Update your Cloud Foundry instance."
}

func (e MinimumUAAAPIVersionNotMetError) Translate(translate func(string, ...interface{}) string) string {
	vars := map[string]interface{}{
		"MinimumVersion": e.MinimumVersion,
	}
	if e.Command != "" {
		vars["Command"] = e.Command
	}
	return translate(e.Error(), vars)
}
