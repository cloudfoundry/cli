package ssh

import (
	"fmt"
	"net"
	"strings"
	"time"

	"code.cloudfoundry.org/clock"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	proxy "github.com/cloudfoundry/socks5-proxy"
	"golang.org/x/crypto/ssh"
)

type Client interface {
	Start() error
	Stop() error

	Dial(net, addr string) (net.Conn, error)
	Listen(net, addr string) (net.Listener, error)
}

type ClientImpl struct {
	opts ClientOpts

	connectionRefusedTimeout time.Duration
	authFailureTimeout       time.Duration
	timeService              clock.Clock
	startDialDelay           time.Duration

	client *ssh.Client

	logTag string
	logger boshlog.Logger
}

func (s *ClientImpl) Start() error {
	authMethods := []ssh.AuthMethod{}

	if s.opts.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(s.opts.PrivateKey))
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing private key '%s'", s.opts.PrivateKey)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if s.opts.Password != "" {
		s.logger.Debug(s.logTag, "Adding password auth method to ssh tunnel config")

		keyboardInteractiveChallenge := func(_, _ string, questions []string, _ []bool) ([]string, error) {
			if len(questions) == 0 {
				return []string{}, nil
			}
			return []string{s.opts.Password}, nil
		}
		authMethods = append(authMethods, ssh.KeyboardInteractive(keyboardInteractiveChallenge))
		authMethods = append(authMethods, ssh.Password(s.opts.Password))
	}

	sshConfig := &ssh.ClientConfig{
		User:            s.opts.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	s.logger.Debug(s.logTag, "Dialing remote server at %s:%d", s.opts.Host, s.opts.Port)

	retryStrategy := &ClientConnectRetryStrategy{
		TimeService:              s.timeService,
		ConnectionRefusedTimeout: s.connectionRefusedTimeout,
		AuthFailureTimeout:       s.authFailureTimeout,
	}

	var err error

	for i := 0; ; i++ {
		s.logger.Debug(s.logTag, "Making attempt #%d", i)

		s.client, err = s.newClient("tcp", net.JoinHostPort(s.opts.Host, fmt.Sprintf("%d", s.opts.Port)), sshConfig)
		if err == nil {
			break
		}

		if !retryStrategy.IsRetryable(err) {
			return bosherr.WrapError(err, "Failed to connect to remote server")
		}

		s.logger.Debug(s.logTag, "Attempt failed #%d: Dialing remote server: %s", i, err.Error())

		time.Sleep(s.startDialDelay)
	}

	return nil
}

func (s *ClientImpl) Dial(n, addr string) (net.Conn, error) {
	return s.client.Dial(n, addr)
}

func (s *ClientImpl) Listen(n, addr string) (net.Listener, error) {
	return s.client.Listen(n, addr)
}

func (s *ClientImpl) Stop() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *ClientImpl) newClient(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	dialFunc := net.Dial

	if !s.opts.DisableSOCKS {
		socksProxy := proxy.NewSocks5Proxy(proxy.NewHostKeyGetter())
		dialFunc = boshhttp.SOCKS5DialFuncFromEnvironment(net.Dial, socksProxy)
	}

	conn, err := dialFunc(network, addr)
	if err != nil {
		return nil, err
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}

	return ssh.NewClient(c, chans, reqs), nil
}

type ClientConnectRetryStrategy struct {
	ConnectionRefusedTimeout time.Duration
	AuthFailureTimeout       time.Duration
	TimeService              clock.Clock

	initialized   bool
	startTime     time.Time
	authStartTime time.Time
}

func (s *ClientConnectRetryStrategy) IsRetryable(err error) bool {
	now := s.TimeService.Now()

	if !s.initialized {
		s.startTime = now
		s.authStartTime = now
		s.initialized = true
	}

	if strings.Contains(err.Error(), "no common algorithms") {
		return false
	}

	if strings.Contains(err.Error(), "unable to authenticate") {
		return now.Before(s.authStartTime.Add(s.AuthFailureTimeout))
	}

	s.authStartTime = now

	return now.Before(s.startTime.Add(s.ConnectionRefusedTimeout))
}
