package cf

type EndpointType string

const (
	UaaEndpointKey             EndpointType = "uaa"
	LoggregatorEndpointKey                  = "loggregator"
	CloudControllerEndpointKey              = "cloud_controller"
)
