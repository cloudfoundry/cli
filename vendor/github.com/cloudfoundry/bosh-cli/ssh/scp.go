package ssh

import (
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

type SCPRunnerImpl struct {
	comboRunner ComboRunner
}

func NewSCPRunner(comboRunner ComboRunner) SCPRunnerImpl {
	return SCPRunnerImpl{comboRunner}
}

func (r SCPRunnerImpl) Run(connOpts ConnectionOpts, result boshdir.SSHResult, scpArgs SCPArgs) error {
	cmdFactory := func(host boshdir.Host, sshArgs SSHArgs) boshsys.Command {
		return boshsys.Command{
			Name: "scp",
			Args: append(sshArgs.OptsForHost(host), scpArgs.ForHost(host)...), // src/dst args come last
		}
	}

	return r.comboRunner.Run(connOpts, result, cmdFactory)
}
