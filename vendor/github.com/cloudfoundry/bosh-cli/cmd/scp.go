package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshssh "github.com/cloudfoundry/bosh-cli/ssh"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type SCPCmd struct {
	deployment boshdir.Deployment
	uuidGen    boshuuid.Generator
	scpRunner  boshssh.SCPRunner
	ui         biui.UI
}

func NewSCPCmd(
	deployment boshdir.Deployment,
	uuidGen boshuuid.Generator,
	scpRunner boshssh.SCPRunner,
	ui biui.UI,
) SCPCmd {
	return SCPCmd{
		deployment: deployment,
		uuidGen:    uuidGen,
		scpRunner:  scpRunner,
		ui:         ui,
	}
}

func (c SCPCmd) Run(opts SCPOpts) error {
	scpArgs := boshssh.NewSCPArgs(opts.Args.Paths, opts.Recursive)

	slug, err := scpArgs.AllOrInstanceGroupOrInstanceSlug()
	if err != nil {
		return err
	}

	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts()
	if err != nil {
		return err
	}

	result, err := c.deployment.SetUpSSH(slug, sshOpts)
	if err != nil {
		return err
	}

	defer func() {
		_ = c.deployment.CleanUpSSH(slug, sshOpts)
	}()

	err = c.scpRunner.Run(connOpts, result, scpArgs)
	if err != nil {
		return bosherr.WrapErrorf(err, "Running SCP")
	}

	return nil
}
