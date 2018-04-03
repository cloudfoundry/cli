package ssh

import (
	"time"

	"code.cloudfoundry.org/clock"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type ClientOpts struct {
	Host string
	Port int

	User       string
	Password   string
	PrivateKey string

	DisableSOCKS bool
}

type ClientFactory struct {
	logger boshlog.Logger
}

func NewClientFactory(logger boshlog.Logger) ClientFactory {
	return ClientFactory{logger: logger}
}

func (f ClientFactory) New(opts ClientOpts) Client {
	return &ClientImpl{
		opts: opts,

		connectionRefusedTimeout: 5 * time.Minute,
		authFailureTimeout:       2 * time.Minute,
		startDialDelay:           500 * time.Millisecond,
		timeService:              clock.NewClock(),

		logTag: "ssh.Client",
		logger: f.logger,
	}
}
