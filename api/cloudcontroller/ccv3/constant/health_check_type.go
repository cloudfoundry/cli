package constant

// HealthCheckType is the manner in which Cloud Foundry verifies the app's status.
type HealthCheckType string

const (
	// HTTP means that CF will make a GET request to the configured HTTP endpoint on the app's default port. Useful if the app can provide an HTTP 200 response.
	HTTP HealthCheckType = "http"
	// Port means that CF will make a TCP connection to the port of ports configured. Useful if the app can receive TCP connections.
	Port HealthCheckType = "port"
	// Process means that Diego ensures that the process(es) stay running. Useful if the app cannot support TCP connections.
	Process HealthCheckType = "process"
)
