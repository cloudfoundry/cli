package constant

// ApplicationHealthCheckType is the method to reach the applications health check
type ApplicationHealthCheckType string

const (
	ApplicationHealthCheckPort    ApplicationHealthCheckType = "port"
	ApplicationHealthCheckHTTP    ApplicationHealthCheckType = "http"
	ApplicationHealthCheckProcess ApplicationHealthCheckType = "process"
)
