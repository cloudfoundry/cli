package ssh

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

type NonInteractiveRunner struct {
	comboRunner ComboRunner
}

func NewNonInteractiveRunner(comboRunner ComboRunner) NonInteractiveRunner {
	return NonInteractiveRunner{comboRunner}
}

func (r NonInteractiveRunner) Run(connOpts ConnectionOpts, result boshdir.SSHResult, rawCmd []string) error {
	if len(result.Hosts) == 0 {
		return bosherr.Errorf("Non-interactive SSH expects at least one host")
	}

	if len(rawCmd) == 0 {
		return bosherr.Errorf("Non-interactive SSH expects non-empty command")
	}

	cmdFactory := func(host boshdir.Host, sshArgs SSHArgs) boshsys.Command {
		return boshsys.Command{
			Name: "ssh",
			Args: append(append(sshArgs.OptsForHost(host), sshArgs.LoginForHost(host)...), rawCmd...),
		}
	}

	return r.comboRunner.Run(connOpts, result, cmdFactory)
}
