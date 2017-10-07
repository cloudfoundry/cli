package clissh

import "net"

type listenerFactory struct{}

func DefaultListenerFactory() listenerFactory {
	return listenerFactory{}
}

func (listenerFactory) Listen(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}
