package actionerror

// ActiveDeploymentNotFoundError is an error wrapper that represents the case
// when the deployment is not found.
type ActiveDeploymentNotFoundError struct {
}

// Error method to display the error message.
func (e ActiveDeploymentNotFoundError) Error() string {
	return "No active deployment found for app."
}
