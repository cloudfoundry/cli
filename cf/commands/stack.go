package commands

import (
	"github.com/cloudfoundry/cli/cf/api/stacks"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ListStack struct {
	ui         terminal.UI
	config     core_config.Reader
	stacksRepo stacks.StackRepository
}

func NewListStack(ui terminal.UI, config core_config.Reader, stacksRepo stacks.StackRepository) (cmd ListStack) {
	cmd.ui = ui
	cmd.config = config
	cmd.stacksRepo = stacksRepo
	return
}

func (cmd ListStack) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "stack",
		Description: T("Show information for a stack (a stack is a pre-built file system, including an operating system, that can run apps)"),
		Usage:       T("CF_NAME stack STACK_NAME"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given stack's guid. All other output for the stack is suppressed.")},
		},
		TotalArgs: 1,
	}
}

func (cmd ListStack) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd ListStack) Run(c *cli.Context) {
	stackName := c.Args()[0]

	stack, apiErr := cmd.stacksRepo.FindByName(stackName)

	if c.Bool("guid") {
		cmd.ui.Say(stack.Guid)
	} else {
		if apiErr != nil {
			cmd.ui.Failed(apiErr.Error())
			return
		}

		cmd.ui.Say(T("Getting stack '{{.Stack}}' in org {{.OrganizationName}} / space {{.SpaceName}} as {{.Username}}...",
			map[string]interface{}{"Stack": stackName,
				"OrganizationName": terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName":        terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"Username":         terminal.EntityNameColor(cmd.config.Username())}))

		cmd.ui.Ok()
		cmd.ui.Say("")

		table := terminal.NewTable(cmd.ui, []string{T("name"), T("description")})
		table.Add(stack.Name, stack.Description)
		table.Print()
	}
}
