package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type UpdateCPIConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewUpdateCPIConfigCmd(ui boshui.UI, director boshdir.Director) UpdateCPIConfigCmd {
	return UpdateCPIConfigCmd{ui: ui, director: director}
}

func (c UpdateCPIConfigCmd) Run(opts UpdateCPIConfigOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.CPIConfig.Bytes)

	bytes, err := tpl.Evaluate(opts.VarFlags.AsVariables(), opts.OpsFlags.AsOp(), boshtpl.EvaluateOpts{})
	if err != nil {
		return bosherr.WrapErrorf(err, "Evaluating cpi config")
	}

	configDiff, err := c.director.DiffCPIConfig(bytes, opts.NoRedact)
	if err != nil {
		return err
	}

	diff := NewDiff(configDiff.Diff)
	diff.Print(c.ui)

	err = c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.director.UpdateCPIConfig(bytes)
}
