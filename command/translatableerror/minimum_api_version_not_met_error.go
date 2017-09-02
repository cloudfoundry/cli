package translatableerror

type MinimumAPIVersionNotMetError struct {
	Command        string
	CurrentVersion string
	MinimumVersion string
}

func (MinimumAPIVersionNotMetError) Error() string {
	return "{{.Command}} requires CF API version {{.MinimumVersion}} or higher. Your target is {{.CurrentVersion}}."
}

func (e MinimumAPIVersionNotMetError) Translate(translate func(string, ...interface{}) string) string {
	if e.Command == "" {
		e.Command = "This command"
	}
	return translate(e.Error(), map[string]interface{}{
		"Command":        e.Command,
		"CurrentVersion": e.CurrentVersion,
		"MinimumVersion": e.MinimumVersion,
	})
}
