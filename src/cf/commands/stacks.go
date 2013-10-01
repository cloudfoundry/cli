package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type Stacks struct {
	ui         terminal.UI
	stacksRepo api.StackRepository
}

func NewStacks(ui terminal.UI, stacksRepo api.StackRepository) (cmd *Stacks) {
	cmd = new(Stacks)
	cmd.ui = ui
	cmd.stacksRepo = stacksRepo
	return
}

func (cmd *Stacks) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd *Stacks) Run(c *cli.Context) {
	cmd.ui.Say("Getting stacks")

	stacks, apiStatus := cmd.stacksRepo.FindAll()
	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	cmd.ui.Ok()

	table := [][]string{
		[]string{"name", "description"},
	}

	for _, stack := range stacks {
		table = append(table, []string{
			stack.Name,
			stack.Description,
		})
	}

	cmd.ui.DisplayTable(table, nil)
}
