package cmd

import (
	"strings"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type LocksCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewLocksCmd(ui boshui.UI, director boshdir.Director) LocksCmd {
	return LocksCmd{ui: ui, director: director}
}

func (c LocksCmd) Run() error {
	locks, err := c.director.Locks()
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "locks",
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Type"),
			boshtbl.NewHeader("Resource"),
			boshtbl.NewHeader("Task ID"),
			boshtbl.NewHeader("Expires at"),
		},
		SortBy: []boshtbl.ColumnSort{{Column: 2, Asc: true}},
	}

	for _, l := range locks {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(l.Type),
			boshtbl.NewValueString(strings.Join(l.Resource, ":")),
			boshtbl.NewValueString(l.TaskID),
			boshtbl.NewValueTime(l.ExpiresAt),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
