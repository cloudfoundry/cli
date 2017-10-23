package actionerror

type MissingNameError struct{}

func (MissingNameError) Error() string {
	return "name not specified for app"
}
