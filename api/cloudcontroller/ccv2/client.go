package ccv2

type Warnings []string

//go:generate counterfeiter . Connection

// Connection creates and executes http requests
type Connection interface {
	Make(passedRequest Request, passedResponse *Response) error
}

//go:generate counterfeiter . ConnectionWrapper

// ConnectionWrapper can wrap a given connection allowing the wrapper to modify
// all requests going in and out of the given connection.
type ConnectionWrapper interface {
	Connection
	Wrap(innerconnection Connection) Connection
}

type CloudControllerClient struct {
	authorizationEndpoint     string
	cloudControllerAPIVersion string
	cloudControllerURL        string
	dopplerEndpoint           string
	loggregatorEndpoint       string
	routingEndpoint           string
	tokenEndpoint             string

	connection Connection
}

// NewCloudControllerClient returns a new CloudControllerClient
func NewCloudControllerClient() *CloudControllerClient {
	return new(CloudControllerClient)
}

// WrapConnection wraps the current CloudControllerClient connection in the
// wrapper
func (client *CloudControllerClient) WrapConnection(wrapper ConnectionWrapper) {
	client.connection = wrapper.Wrap(client.connection)
}
