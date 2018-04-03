package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type EventsCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewEventsCmd(ui boshui.UI, director boshdir.Director) EventsCmd {
	return EventsCmd{ui: ui, director: director}
}

func (c EventsCmd) Run(opts EventsOpts) error {
	filter := boshdir.EventsFilter{
		BeforeID:   opts.BeforeID,
		Before:     opts.Before,
		After:      opts.After,
		Deployment: opts.Deployment,
		Task:       opts.Task,
		Instance:   opts.Instance,
		User:       opts.User,
		Action:     opts.Action,
		ObjectType: opts.ObjectType,
		ObjectName: opts.ObjectName,
	}

	events, err := c.director.Events(filter)
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "events",
		Header: []boshtbl.Header{
			boshtbl.NewHeader("ID"),
			boshtbl.NewHeader("Time"),
			boshtbl.NewHeader("User"),
			boshtbl.NewHeader("Action"),
			boshtbl.NewHeader("Object Type"),
			boshtbl.NewHeader("Object Name"),
			boshtbl.NewHeader("Task ID"),
			boshtbl.NewHeader("Deployment"),
			boshtbl.NewHeader("Instance"),
			boshtbl.NewHeader("Context"),
			boshtbl.NewHeader("Error"),
		},
	}

	for _, e := range events {
		id := e.ID()

		if e.ParentID() != "" {
			id += " <- " + e.ParentID()
		}

		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(id),
			boshtbl.NewValueTime(e.Timestamp()),
			boshtbl.NewValueString(e.User()),
			boshtbl.NewValueString(e.Action()),
			boshtbl.NewValueString(e.ObjectType()),
			boshtbl.NewValueString(e.ObjectName()),
			boshtbl.NewValueString(e.TaskID()),
			boshtbl.NewValueString(e.DeploymentName()),
			boshtbl.NewValueString(e.Instance()),
			boshtbl.NewValueInterface(e.Context()),
			boshtbl.NewValueString(e.Error()),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
