package translatableerror

type MinimumAPIVersionNotMetError struct {
	CurrentVersion string
	MinimumVersion string
}

func (MinimumAPIVersionNotMetError) Error() string {
	return "This command requires CF API version {{.MinimumVersion}} or higher. Your target is {{.CurrentVersion}}."
}

func (e MinimumAPIVersionNotMetError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"CurrentVersion": e.CurrentVersion,
		"MinimumVersion": e.MinimumVersion,
	})
}
