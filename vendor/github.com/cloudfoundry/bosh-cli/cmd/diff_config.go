package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type DiffConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewDiffConfigCmd(ui boshui.UI, director boshdir.Director) DiffConfigCmd {
	return DiffConfigCmd{ui: ui, director: director}
}

func (c DiffConfigCmd) Run(opts DiffConfigOpts) error {

	fromID := opts.Args.FromID
	toID := opts.Args.ToID

	configDiff, err := c.director.DiffConfigByID(fromID, toID)
	if err != nil {
		return err
	}

	diff := NewDiff(configDiff.Diff)
	diff.Print(c.ui)

	return nil
}
