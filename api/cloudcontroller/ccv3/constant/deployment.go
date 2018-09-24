package constant

// A zero down time deployment used to push new apps without restart
type DeploymentState string

const (
	// Deployment is in state 'DEPLOYING'
	DeploymentDeploying DeploymentState = "DEPLOYING"

	// Deployment is in state 'CANCELED'
	DeploymentCanceled DeploymentState = "CANCELED"

	// Deployment is in state 'DEPLOYED'
	DeploymentDeployed DeploymentState = "DEPLOYED"
)
