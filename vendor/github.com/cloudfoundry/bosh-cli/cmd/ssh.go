package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshssh "github.com/cloudfoundry/bosh-cli/ssh"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type SSHCmd struct {
	deployment       boshdir.Deployment
	uuidGen          boshuuid.Generator
	intSSHRunner     boshssh.Runner
	nonIntSSHRunner  boshssh.Runner
	resultsSSHRunner boshssh.Runner
	ui               boshui.UI
}

func NewSSHCmd(
	deployment boshdir.Deployment,
	uuidGen boshuuid.Generator,
	intSSHRunner boshssh.Runner,
	nonIntSSHRunner boshssh.Runner,
	resultsSSHRunner boshssh.Runner,
	ui boshui.UI,
) SSHCmd {
	return SSHCmd{
		deployment:       deployment,
		uuidGen:          uuidGen,
		intSSHRunner:     intSSHRunner,
		nonIntSSHRunner:  nonIntSSHRunner,
		resultsSSHRunner: resultsSSHRunner,
		ui:               ui,
	}
}

func (c SSHCmd) Run(opts SSHOpts) error {
	if opts.Results || !c.ui.IsInteractive() {
		if len(opts.Command) == 0 {
			return bosherr.Errorf("Non-interactive SSH requires non-empty command")
		}
	}

	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts()
	if err != nil {
		return err
	}

	connOpts.RawOpts = opts.RawOpts.AsStrings()

	result, err := c.deployment.SetUpSSH(opts.Args.Slug, sshOpts)
	if err != nil {
		return err
	}

	defer func() {
		_ = c.deployment.CleanUpSSH(opts.Args.Slug, sshOpts)
	}()

	var runner boshssh.Runner

	if opts.Results {
		runner = c.resultsSSHRunner
	} else if !c.ui.IsInteractive() || len(opts.Command) > 0 {
		runner = c.nonIntSSHRunner
	} else {
		runner = c.intSSHRunner
	}

	err = runner.Run(connOpts, result, opts.Command)
	if err != nil {
		return bosherr.WrapErrorf(err, "Running SSH")
	}

	return nil
}
