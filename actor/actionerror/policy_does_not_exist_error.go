package actionerror

type PolicyDoesNotExistError struct{}

func (PolicyDoesNotExistError) Error() string {
	return "policy does not exist."
}
