package constant

// DeploymentStrategy is the strategy used to push an application.
type DeploymentStrategy string

const (
	// Default means app will be stopped and started with new droplet.
	DeploymentStrategyDefault DeploymentStrategy = ""

	// Rolling means a new web process will be created for the app and instances will roll from the old one to the new one.
	DeploymentStrategyRolling DeploymentStrategy = "rolling"
)
