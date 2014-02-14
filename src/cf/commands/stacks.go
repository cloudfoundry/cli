package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListStacks struct {
	ui         terminal.UI
	config     configuration.Reader
	stacksRepo api.StackRepository
}

func NewListStacks(ui terminal.UI, config configuration.Reader, stacksRepo api.StackRepository) (cmd ListStacks) {
	cmd.ui = ui
	cmd.config = config
	cmd.stacksRepo = stacksRepo
	return
}

func (cmd ListStacks) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = append(reqs, reqFactory.NewLoginRequirement())
	return
}

func (cmd ListStacks) Run(c *cli.Context) {
	cmd.ui.Say("Getting stacks in org %s / space %s as %s...",
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	stacks, apiResponse := cmd.stacksRepo.FindAll()
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table := [][]string{
		[]string{"name", "description"},
	}

	for _, stack := range stacks {
		table = append(table, []string{
			stack.Name,
			stack.Description,
		})
	}

	cmd.ui.DisplayTable(table)
}
