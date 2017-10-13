package clissh

import (
	"golang.org/x/crypto/ssh"
)

type secureDialer struct{}

func DefaultSecureDialer() secureDialer {
	return secureDialer{}
}

func (secureDialer) Dial(network string, address string, config *ssh.ClientConfig) (SecureClient, error) {
	client, err := ssh.Dial(network, address, config)
	if err != nil {
		return secureClient{}, err
	}

	return secureClient{client: client}, nil
}
