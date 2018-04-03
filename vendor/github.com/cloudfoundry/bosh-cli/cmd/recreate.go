package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type RecreateCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewRecreateCmd(ui boshui.UI, deployment boshdir.Deployment) RecreateCmd {
	return RecreateCmd{ui: ui, deployment: deployment}
}

func (c RecreateCmd) Run(opts RecreateOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	recreateOpts := boshdir.RecreateOpts{
		SkipDrain:   opts.SkipDrain,
		Force:       opts.Force,
		Fix:         opts.Fix,
		DryRun:      opts.DryRun,
		Canaries:    opts.Canaries,
		MaxInFlight: opts.MaxInFlight,
	}

	return c.deployment.Recreate(opts.Args.Slug, recreateOpts)
}
