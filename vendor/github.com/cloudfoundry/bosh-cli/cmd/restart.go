package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type RestartCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewRestartCmd(ui boshui.UI, deployment boshdir.Deployment) RestartCmd {
	return RestartCmd{ui: ui, deployment: deployment}
}

func (c RestartCmd) Run(opts RestartOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	restartOpts := boshdir.RestartOpts{
		SkipDrain:   opts.SkipDrain,
		Force:       opts.Force,
		Canaries:    opts.Canaries,
		MaxInFlight: opts.MaxInFlight,
	}
	return c.deployment.Restart(opts.Args.Slug, restartOpts)
}
