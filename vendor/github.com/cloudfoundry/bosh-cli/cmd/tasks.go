package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type TasksCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewTasksCmd(ui boshui.UI, director boshdir.Director) TasksCmd {
	return TasksCmd{ui: ui, director: director}
}

func (c TasksCmd) Run(opts TasksOpts) error {
	filter := boshdir.TasksFilter{
		All:        opts.All,
		Deployment: opts.Deployment,
	}

	if opts.Recent != nil {
		return c.printTable(c.director.RecentTasks(*opts.Recent, filter))
	}

	filter.All = true
	return c.printTable(c.director.CurrentTasks(filter))
}

func (c TasksCmd) printTable(tasks []boshdir.Task, err error) error {
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "tasks",
		Header: []boshtbl.Header{
			boshtbl.NewHeader("ID"),
			boshtbl.NewHeader("State"),
			boshtbl.NewHeader("Started At"),
			boshtbl.NewHeader("Last Activity At"),
			boshtbl.NewHeader("User"),
			boshtbl.NewHeader("Deployment"),
			boshtbl.NewHeader("Description"),
			boshtbl.NewHeader("Result"),
		},
		SortBy: []boshtbl.ColumnSort{{Column: 0}},
	}

	for _, t := range tasks {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueInt(t.ID()),
			boshtbl.ValueFmt{
				V:     boshtbl.NewValueString(t.State()),
				Error: t.IsError(),
			},
			boshtbl.NewValueTime(t.StartedAt()),
			boshtbl.NewValueTime(t.LastActivityAt()),
			boshtbl.NewValueString(t.User()),
			boshtbl.NewValueString(t.DeploymentName()),
			boshtbl.NewValueString(t.Description()),
			boshtbl.NewValueString(t.Result()),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
