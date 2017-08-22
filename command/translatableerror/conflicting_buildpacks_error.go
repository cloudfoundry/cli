package translatableerror

type ConflictingBuildpacksError struct {
}

func (ConflictingBuildpacksError) Error() string {
	return "Cannot specify 'null' or 'default' with other buildpacks"
}

func (e ConflictingBuildpacksError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), nil)
}
