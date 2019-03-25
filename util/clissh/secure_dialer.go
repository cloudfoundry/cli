package clissh

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

//go:generate counterfeiter . SecureDialer

type SecureDialer interface {
	Dial(network, address string, config *ssh.ClientConfig) (SecureClient, error)
}

type fakeConn struct{}

type secureDialer struct{}

func DefaultSecureDialer() secureDialer {
	return secureDialer{}
}

func (secureDialer) Dial(network string, address string, config *ssh.ClientConfig) (SecureClient, error) {
	conn, err := proxy.FromEnvironment().Dial(network, address)
	if err != nil {
		return secureClient{}, err
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, address, config)
	if err != nil {
		return secureClient{}, err
	}
	client := ssh.NewClient(c, chans, reqs)

	return secureClient{client: client}, nil
}
