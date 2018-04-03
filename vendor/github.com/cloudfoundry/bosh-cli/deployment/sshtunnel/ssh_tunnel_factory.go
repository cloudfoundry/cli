package sshtunnel

import (
	boshssh "github.com/cloudfoundry/bosh-cli/ssh"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Options struct {
	Host string
	Port int

	User       string
	Password   string
	PrivateKey string

	LocalForwardPort  int
	RemoteForwardPort int
}

func (o Options) IsEmpty() bool {
	return o == Options{}
}

type Factory interface {
	NewSSHTunnel(Options) SSHTunnel
}

type factory struct {
	logger boshlog.Logger
}

func NewFactory(logger boshlog.Logger) Factory {
	return &factory{logger: logger}
}

func (f *factory) NewSSHTunnel(opts Options) SSHTunnel {
	clientFactory := boshssh.NewClientFactory(f.logger)

	clientOpts := boshssh.ClientOpts{
		Host: opts.Host,
		Port: opts.Port,

		User:       opts.User,
		Password:   opts.Password,
		PrivateKey: opts.PrivateKey,
	}

	tunnel := &sshTunnel{
		client: clientFactory.New(clientOpts),

		localForwardPort:  opts.LocalForwardPort,
		remoteForwardPort: opts.RemoteForwardPort,

		logTag: "sshTunnel",
		logger: f.logger,
	}

	return tunnel
}
