package ccerror

type InvalidBuildpackError struct {
}

func (e InvalidBuildpackError) Error() string {
	return "Buildpack must be an existing admin buildpack or a valid git URI"
}
