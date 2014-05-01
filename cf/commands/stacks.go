package commands

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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

func (command ListStacks) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "stacks",
		Description: "List all stacks (a stack is a pre-built file system, including an operating system, that can run apps)",
		Usage:       "CF_NAME stacks",
	}
}

func (cmd ListStacks) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd ListStacks) Run(c *cli.Context) {
	cmd.ui.Say("Getting stacks in org %s / space %s as %s...",
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	stacks, apiErr := cmd.stacksRepo.FindAll()
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table := terminal.NewTable(cmd.ui, []string{"name", "description"})

	for _, stack := range stacks {
		table.Add([]string{
			stack.Name,
			stack.Description,
		})
	}

	table.Print()
}
