package constant

// ApplicationHealthCheckType is the method to reach the applications health check
type ApplicationHealthCheckType string

const (
	ApplicationHealthCheckHTTP    ApplicationHealthCheckType = "http"
	ApplicationHealthCheckPort    ApplicationHealthCheckType = "port"
	ApplicationHealthCheckProcess ApplicationHealthCheckType = "process"
)

// ApplicationPackageState is the staging state of application bits.
type ApplicationPackageState string

const (
	ApplicationPackageFailed  ApplicationPackageState = "FAILED"
	ApplicationPackagePending ApplicationPackageState = "PENDING"
	ApplicationPackageStaged  ApplicationPackageState = "STAGED"
	ApplicationPackageUnknown ApplicationPackageState = "UNKNOWN"
)

// ApplicationState is the running state of an application.
type ApplicationState string

const (
	ApplicationStarted ApplicationState = "STARTED"
	ApplicationStopped ApplicationState = "STOPPED"
)
