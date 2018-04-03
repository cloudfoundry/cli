package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type DeleteSnapshotsCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewDeleteSnapshotsCmd(ui boshui.UI, deployment boshdir.Deployment) DeleteSnapshotsCmd {
	return DeleteSnapshotsCmd{ui: ui, deployment: deployment}
}

func (c DeleteSnapshotsCmd) Run() error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.deployment.DeleteSnapshots()
}
