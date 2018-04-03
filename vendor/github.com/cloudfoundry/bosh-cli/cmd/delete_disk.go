package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type DeleteDiskCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewDeleteDiskCmd(ui boshui.UI, director boshdir.Director) DeleteDiskCmd {
	return DeleteDiskCmd{ui: ui, director: director}
}

func (c DeleteDiskCmd) Run(opts DeleteDiskOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	disk, err := c.director.FindOrphanDisk(opts.Args.CID)
	if err != nil {
		return err
	}

	return disk.Delete()
}
