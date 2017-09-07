package translatableerror

type MinimumAPIVersionNotMetError struct {
	Command        string
	CurrentVersion string
	MinimumVersion string
}

func (e MinimumAPIVersionNotMetError) Error() string {
	switch {
	case e.Command != "" && e.CurrentVersion != "":
		return "{{.Command}} requires CF API version {{.MinimumVersion}} or higher. Your target is {{.CurrentVersion}}."
	case e.Command != "" && e.CurrentVersion == "":
		return "{{.Command}} requires CF API version {{.MinimumVersion}} or higher."
	case e.Command == "" && e.CurrentVersion != "":
		return "This command requires CF API version {{.MinimumVersion}} or higher. Your target is {{.CurrentVersion}}."
	default:
		return "This command requires CF API version {{.MinimumVersion}} or higher."
	}
}

func (e MinimumAPIVersionNotMetError) Translate(translate func(string, ...interface{}) string) string {
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
