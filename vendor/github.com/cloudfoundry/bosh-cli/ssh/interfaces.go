package ssh

import (
	"io"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

//go:generate counterfeiter . Runner

type Runner interface {
	Run(ConnectionOpts, boshdir.SSHResult, []string) error
}

//go:generate counterfeiter . SCPRunner

type SCPRunner interface {
	Run(ConnectionOpts, boshdir.SSHResult, SCPArgs) error
}

type ConnectionOpts struct {
	PrivateKey string

	GatewayDisable bool

	GatewayUsername       string
	GatewayHost           string
	GatewayPrivateKeyPath string

	SOCKS5Proxy string

	RawOpts []string
}

//go:generate counterfeiter . Session

type Session interface {
	Start() (SSHArgs, error)
	Finish() error
}

type Writer interface {
	ForInstance(jobName, indexOrID string) InstanceWriter
	Flush()
}

type InstanceWriter interface {
	Stdout() io.Writer
	Stderr() io.Writer
	End(exitStatus int, err error)
}
