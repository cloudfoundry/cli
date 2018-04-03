package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type VariablesCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewVariablesCmd(ui boshui.UI, deployment boshdir.Deployment) VariablesCmd {
	return VariablesCmd{ui: ui, deployment: deployment}
}

func (c VariablesCmd) Run() error {

	variables, err := c.deployment.Variables()
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "variables",

		Header: []boshtbl.Header{
			boshtbl.NewHeader("ID"),
			boshtbl.NewHeader("Name"),
		},

		SortBy: []boshtbl.ColumnSort{
			{Column: 1, Asc: true},
		},
	}

	for _, variable := range variables {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(variable.ID),
			boshtbl.NewValueString(variable.Name),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
