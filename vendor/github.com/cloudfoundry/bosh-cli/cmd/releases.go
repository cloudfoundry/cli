package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type ReleasesCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewReleasesCmd(ui boshui.UI, director boshdir.Director) ReleasesCmd {
	return ReleasesCmd{ui: ui, director: director}
}

func (c ReleasesCmd) Run() error {
	releases, err := c.director.Releases()
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "releases",

		Header: []boshtbl.Header{
			boshtbl.NewHeader("Name"),
			boshtbl.NewHeader("Version"),
			boshtbl.NewHeader("Commit Hash"),
		},

		SortBy: []boshtbl.ColumnSort{
			{Column: 0, Asc: true},
			{Column: 1},
		},

		Notes: []string{
			"(*) Currently deployed",
			"(+) Uncommitted changes",
		},
	}

	for _, rel := range releases {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(rel.Name()),
			boshtbl.NewValueSuffix(
				boshtbl.NewValueVersion(rel.Version()),
				rel.VersionMark("*"),
			),
			boshtbl.NewValueString(rel.CommitHashWithMark("+")),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
