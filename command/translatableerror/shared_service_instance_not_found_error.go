package translatableerror

type SharedServiceInstanceNotFoundError struct {
}

func (SharedServiceInstanceNotFoundError) Error() string {
	return "Specified instance not found or not a managed service instance. Sharing is not supported for user provided services."
}

func (e SharedServiceInstanceNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
