package test_helpers

import (
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.org/x/crypto/ssh"
)

func WaitFor(f func() error) error {
	ch := make(chan error)
	go func() {
		err := f()
		ch <- err
	}()
	var err error
	Eventually(ch, 10).Should(Receive(&err))
	return err
}

func Pipe() (net.Conn, net.Conn) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	Expect(err).NotTo(HaveOccurred())

	address := listener.Addr().String()

	serverConnCh := make(chan net.Conn, 1)
	go func(serverConnCh chan net.Conn, listener net.Listener) {
		defer GinkgoRecover()
		conn, err := listener.Accept()
		Expect(err).NotTo(HaveOccurred())

		serverConnCh <- conn
	}(serverConnCh, listener)

	clientConn, err := net.Dial("tcp", address)
	Expect(err).NotTo(HaveOccurred())

	return <-serverConnCh, clientConn
}

func NewClient(clientNetConn net.Conn, clientConfig *ssh.ClientConfig) *ssh.Client {
	if clientConfig == nil {
		clientConfig = &ssh.ClientConfig{
			User: "username",
			Auth: []ssh.AuthMethod{
				ssh.Password("secret"),
			},
		}
	}

	clientConn, clientChannels, clientRequests, clientConnErr := ssh.NewClientConn(clientNetConn, "0.0.0.0", clientConfig)
	Expect(clientConnErr).NotTo(HaveOccurred())

	return ssh.NewClient(clientConn, clientChannels, clientRequests)
}

type TestNetError struct {
	timeout   bool
	temporary bool
}

func NewTestNetError(timeout, temporary bool) *TestNetError {
	return &TestNetError{
		timeout:   timeout,
		temporary: temporary,
	}
}

func (e *TestNetError) Error() string   { return "test error" }
func (e *TestNetError) Timeout() bool   { return e.timeout }
func (e *TestNetError) Temporary() bool { return e.temporary }
