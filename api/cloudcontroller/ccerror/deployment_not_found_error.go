package ccerror

// DeploymentNotFoundError is returned when an endpoint cannot find the
// specified deployment
type DeploymentNotFoundError struct {
}

func (e DeploymentNotFoundError) Error() string {
	return "Deployment not found"
}
