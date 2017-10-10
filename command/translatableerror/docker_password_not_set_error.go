package translatableerror

type DockerPasswordNotSetError struct{}

func (DockerPasswordNotSetError) Error() string {
	return "Environment variable CF_DOCKER_PASSWORD not set."
}

func (e DockerPasswordNotSetError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
