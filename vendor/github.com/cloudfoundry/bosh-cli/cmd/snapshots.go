package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type SnapshotsCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewSnapshotsCmd(ui boshui.UI, deployment boshdir.Deployment) SnapshotsCmd {
	return SnapshotsCmd{ui: ui, deployment: deployment}
}

func (c SnapshotsCmd) Run(opts SnapshotsOpts) error {
	snapshots, err := c.deployment.Snapshots()
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "snapshots",
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Instance"),
			boshtbl.NewHeader("CID"),
			boshtbl.NewHeader("Created At"),
			boshtbl.NewHeader("Clean"),
		},
	}

	for _, s := range snapshots {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(s.InstanceDesc()),
			boshtbl.NewValueString(s.CID),
			boshtbl.NewValueTime(s.CreatedAt),
			boshtbl.NewValueBool(s.Clean),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
