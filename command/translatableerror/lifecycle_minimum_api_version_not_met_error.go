package translatableerror

type LifecycleMinimumAPIVersionNotMetError struct {
	CurrentVersion string
	MinimumVersion string
}

func (LifecycleMinimumAPIVersionNotMetError) Error() string {
	return "Lifecycle value 'staging' requires CF API version {{.MinimumVersion}} or higher. Your target is {{.CurrentVersion}}."
}

func (e LifecycleMinimumAPIVersionNotMetError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"CurrentVersion": e.CurrentVersion,
		"MinimumVersion": e.MinimumVersion,
	})
}
