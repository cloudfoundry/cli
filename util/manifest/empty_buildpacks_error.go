package manifest

type EmptyBuildpacksError struct{}

func (EmptyBuildpacksError) Error() string {
	return "Buildpacks property cannot be an empty string."
}
