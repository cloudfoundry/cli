package clissh

import (
	"net"

	"golang.org/x/crypto/ssh"
)

//go:generate counterfeiter . SecureClient

type SecureClient interface {
	NewSession() (SecureSession, error)
	Conn() ssh.Conn
	Dial(network, address string) (net.Conn, error)
	Wait() error
	Close() error
}

type secureClient struct {
	client *ssh.Client
}

func (sc secureClient) Close() error {
	return sc.client.Close()
}

func (sc secureClient) Conn() ssh.Conn {
	return sc.client.Conn
}

func (sc secureClient) Dial(n, addr string) (net.Conn, error) {
	return sc.client.Dial(n, addr)
}

func (sc secureClient) NewSession() (SecureSession, error) {
	return sc.client.NewSession()
}

func (sc secureClient) Wait() error {
	return sc.client.Wait()
}
