package translatableerror

type InvalidBuildpacksError struct{}

func (InvalidBuildpacksError) Error() string {
	return "Multiple buildpacks flags cannot have null/default option."
}

func (e InvalidBuildpacksError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
