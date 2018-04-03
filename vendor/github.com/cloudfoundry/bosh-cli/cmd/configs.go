package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type ConfigsCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewConfigsCmd(ui boshui.UI, director boshdir.Director) ConfigsCmd {
	return ConfigsCmd{ui: ui, director: director}
}

func (c ConfigsCmd) Run(opts ConfigsOpts) error {
	filter := boshdir.ConfigsFilter{
		Type: opts.Type,
		Name: opts.Name,
	}

	configs, err := c.director.ListConfigs(opts.Recent, filter)
	if err != nil {
		return err
	}

	var headers []boshtbl.Header
	headers = append(headers, boshtbl.NewHeader("ID"))
	headers = append(headers, boshtbl.NewHeader("Type"))
	headers = append(headers, boshtbl.NewHeader("Name"))
	headers = append(headers, boshtbl.NewHeader("Team"))
	headers = append(headers, boshtbl.NewHeader("Created At"))

	table := boshtbl.Table{
		Content: "configs",
		Header:  headers,
	}

	for _, config := range configs {
		var result []boshtbl.Value
		result = append(result, boshtbl.NewValueString(config.ID))
		result = append(result, boshtbl.NewValueString(config.Type))
		result = append(result, boshtbl.NewValueString(config.Name))
		result = append(result, boshtbl.NewValueString(config.Team))
		result = append(result, boshtbl.NewValueString(config.CreatedAt))
		table.Rows = append(table.Rows, result)
	}

	c.ui.PrintTable(table)
	return nil
}
