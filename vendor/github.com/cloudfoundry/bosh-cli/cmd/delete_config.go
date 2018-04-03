package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DeleteConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewDeleteConfigCmd(ui boshui.UI, director boshdir.Director) DeleteConfigCmd {
	return DeleteConfigCmd{ui: ui, director: director}
}

func (c DeleteConfigCmd) Run(opts DeleteConfigOpts) error {
	if opts == (DeleteConfigOpts{}) {
		return bosherr.Error("Either <ID> or parameters --type and --name must be provided")
	}

	if opts.Args.ID != "" && (opts.Type != "" || opts.Name != "") {
		return bosherr.Error("Can only specify one of ID or parameters '--type' and '--name'")
	}

	if (opts.Args.ID == "" && opts.Type != "" && opts.Name == "") || (opts.Args.ID == "" && opts.Name != "" && opts.Type == "") {
		return bosherr.Error("Need to specify both parameters '--type' and '--name'")
	}

	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	var deleted bool

	if opts.Args.ID != "" {
		deleted, err = c.director.DeleteConfigByID(opts.Args.ID)
		if !deleted {
			c.ui.PrintLinef("No configs to delete: no matches for id '%s' found.", opts.Args.ID)
		}
	} else {
		deleted, err = c.director.DeleteConfig(opts.Type, opts.Name)
		if !deleted {
			c.ui.PrintLinef("No configs to delete: no matches for type '%s' and name '%s' found.", opts.Type, opts.Name)
		}
	}
	if err != nil {
		return err
	}

	return nil
}
