package translatableerror

type EmptyBuildpacksError struct{}

func (EmptyBuildpacksError) Error() string {
	return "Buildpacks property cannot be an empty string."
}

func (e EmptyBuildpacksError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
