package clissh

import "net"

//go:generate counterfeiter . ListenerFactory

type ListenerFactory interface {
	Listen(network, address string) (net.Listener, error)
}

type listenerFactory struct{}

func DefaultListenerFactory() listenerFactory {
	return listenerFactory{}
}

func (listenerFactory) Listen(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}
