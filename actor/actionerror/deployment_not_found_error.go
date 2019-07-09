package actionerror

// DeploymentNotFoundError is an error wrapper that represents the case
// when the deployment is not found.
type DeploymentNotFoundError struct {
}

// Error method to display the error message.
func (e DeploymentNotFoundError) Error() string {
	return "Deployment not found."
}
