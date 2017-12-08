package constant

// ApplicationState represents the current desired state of the app.
type ApplicationState string

const (
	// ApplicationStopped is a desired 'stopped' state.
	ApplicationStopped ApplicationState = "STOPPED"
	// ApplicationStarted is a desired 'started' state.
	ApplicationStarted ApplicationState = "STARTED"
)

// AppLifecycleType informs the platform of how to build droplets and run apps.
type AppLifecycleType string

const (
	// BuildpackAppLifecycleType will use a droplet and a rootfs to run the app.
	BuildpackAppLifecycleType AppLifecycleType = "buildpack"
	// DockerAppLifecycleType will pull a docker image from a registry to run an
	// app.
	DockerAppLifecycleType AppLifecycleType = "docker"
)
