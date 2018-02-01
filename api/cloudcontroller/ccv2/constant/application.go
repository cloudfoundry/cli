package constant

// ApplicationHealthCheckType is the method to reach the applications health check
type ApplicationHealthCheckType string

const (
	ApplicationHealthCheckPort    ApplicationHealthCheckType = "port"
	ApplicationHealthCheckHTTP    ApplicationHealthCheckType = "http"
	ApplicationHealthCheckProcess ApplicationHealthCheckType = "process"
)

// ApplicationState is the running state of an application.
type ApplicationState string

const (
	ApplicationStarted ApplicationState = "STARTED"
	ApplicationStopped ApplicationState = "STOPPED"
)

// ApplicationPackageState is the staging state of application bits.
type ApplicationPackageState string

const (
	ApplicationPackageStaged  ApplicationPackageState = "STAGED"
	ApplicationPackagePending ApplicationPackageState = "PENDING"
	ApplicationPackageFailed  ApplicationPackageState = "FAILED"
	ApplicationPackageUnknown ApplicationPackageState = "UNKNOWN"
)
