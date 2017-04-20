package plugin

//go:generate counterfeiter . ConnectionWrapper

// ConnectionWrapper can wrap a given connection allowing the wrapper to modify
// all requests going in and out of the given connection.
type ConnectionWrapper interface {
	Connection
	Wrap(innerconnection Connection) Connection
}

// WrapConnection wraps the current Client connection in the wrapper.
func (client *Client) WrapConnection(wrapper ConnectionWrapper) {
	client.connection = wrapper.Wrap(client.connection)
}
