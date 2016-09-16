package cloudcontrollerv2

type Warnings []string

type CloudControllerClient struct {
	authorizationEndpoint     string
	cloudControllerAPIVersion string
	cloudControllerURL        string
	dopplerEndpoint           string
	loggregatorEndpoint       string
	routingEndpoint           string
	tokenEndpoint             string

	connection *Connection
}

func NewCloudControllerClient() *CloudControllerClient {
	return new(CloudControllerClient)
}
