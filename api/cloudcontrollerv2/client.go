package cloudcontrollerv2

type Warnings []string

type CloudControllerClient struct {
	cloudControllerURL        string
	cloudControllerAPIVersion string
	authorizationEndpoint     string
	loggregatorEndpoint       string
	dopplerEndpoint           string
	tokenEndpoint             string

	connection *Connection
}

func NewCloudControllerClient() *CloudControllerClient {
	return new(CloudControllerClient)
}
