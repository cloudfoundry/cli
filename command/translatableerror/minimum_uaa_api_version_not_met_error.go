package translatableerror

type MinimumUAAAPIVersionNotMetError struct {
	Command        string
	CurrentVersion string
	MinimumVersion string
}

func (e MinimumUAAAPIVersionNotMetError) Error() string {
	switch {
	case e.Command != "" && e.CurrentVersion != "":
		return "{{.Command}} requires UAA API version {{.MinimumVersion}} or higher, but your current version is {{.CurrentVersion}}"
	case e.Command != "" && e.CurrentVersion == "":
		return "{{.Command}} requires UAA API version {{.MinimumVersion}} or higher."
	case e.Command == "" && e.CurrentVersion != "":
		return "This command requires UAA API version {{.MinimumVersion}} or higher, but your current version is {{.CurrentVersion}}"
	default:
		return "This command requires UAA API version {{.MinimumVersion}} or higher."
	}
}

func (e MinimumUAAAPIVersionNotMetError) Translate(translate func(string, ...interface{}) string) string {
	vars := map[string]interface{}{
		"MinimumVersion": e.MinimumVersion,
	}
	if e.CurrentVersion != "" {
		vars["CurrentVersion"] = e.CurrentVersion
	}
	if e.Command != "" {
		vars["Command"] = e.Command
	}
	return translate(e.Error(), vars)
}
