package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type ErrandsCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewErrandsCmd(ui boshui.UI, deployment boshdir.Deployment) ErrandsCmd {
	return ErrandsCmd{ui: ui, deployment: deployment}
}

func (c ErrandsCmd) Run() error {
	errands, err := c.deployment.Errands()
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "errands",
		Header:  []boshtbl.Header{boshtbl.NewHeader("Name")},
		SortBy:  []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	for _, e := range errands {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(e.Name),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
