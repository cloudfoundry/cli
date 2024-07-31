package constant

// DeploymentStrategy is the strategy used to push an application.
type DeploymentStrategy string

const (
	// Default means app will be stopped and started with new droplet.
	DeploymentStrategyDefault DeploymentStrategy = ""

	// Rolling means a new web process will be created for the app and instances will roll from the old one to the new one.
	DeploymentStrategyRolling DeploymentStrategy = "rolling"

	// Canary means after a web process is created for the app the deployment will pause for evaluation until it is continued or canceled.
	DeploymentStrategyCanary DeploymentStrategy = "canary"
)
