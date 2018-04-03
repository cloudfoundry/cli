package cmd

import (
	"errors"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type DisksCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewDisksCmd(ui boshui.UI, director boshdir.Director) DisksCmd {
	return DisksCmd{ui: ui, director: director}
}

func (c DisksCmd) Run(opts DisksOpts) error {
	if !opts.Orphaned {
		return errors.New("Only --orphaned is supported")
	}

	disks, err := c.director.OrphanDisks()
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "disks",
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Disk CID"),
			boshtbl.NewHeader("Size"),
			boshtbl.NewHeader("Deployment"),
			boshtbl.NewHeader("Instance"),
			boshtbl.NewHeader("AZ"),
			boshtbl.NewHeader("Orphaned At"),
		},
		SortBy: []boshtbl.ColumnSort{{Column: 5}},
	}

	for _, d := range disks {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(d.CID()),
			boshtbl.NewValueMegaBytes(d.Size()),
			boshtbl.NewValueString(d.Deployment().Name()),
			boshtbl.NewValueString(d.InstanceName()),
			boshtbl.NewValueString(d.AZName()),
			boshtbl.NewValueTime(d.OrphanedAt()),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
