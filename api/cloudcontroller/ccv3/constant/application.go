package constant

// AppLifecycleType informs the platform of how to build droplets and run apps.
type AppLifecycleType string

const (
	// AppLifecycleTypeBuildpack will use a droplet and a rootfs to run the app.
	AppLifecycleTypeBuildpack AppLifecycleType = "buildpack"
	// AppLifecycleTypeDocker will pull a docker image from a registry to run an
	// app.
	AppLifecycleTypeDocker AppLifecycleType = "docker"
)

// ApplicationAction represents the action being taken on an application
type ApplicationAction string

const (
	// ApplicationStarting indicates that the app is being started
	ApplicationStarting ApplicationAction = "Starting"
	// ApplicationRestarting indicates that the app is being restarted
	ApplicationRestarting ApplicationAction = "Restarting"
	// ApplicationRollingBack indicates that the app is being rolled back to a
	// prior revision
	ApplicationRollingBack ApplicationAction = "Rolling Back"
)

// ApplicationState represents the current desired state of the app.
type ApplicationState string

const (
	// ApplicationStopped is a desired 'stopped' state.
	ApplicationStopped ApplicationState = "STOPPED"
	// ApplicationStarted is a desired 'started' state.
	ApplicationStarted ApplicationState = "STARTED"
)
