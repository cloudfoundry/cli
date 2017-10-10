package actionerror

type DockerPasswordNotSetError struct {
}

func (DockerPasswordNotSetError) Error() string {
	return "Docker password not set."
}
