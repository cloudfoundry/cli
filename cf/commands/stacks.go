package commands

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
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

func (cmd ListStacks) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "stacks",
		Description: T("List all stacks (a stack is a pre-built file system, including an operating system, that can run apps)"),
		Usage:       T("CF_NAME stacks"),
	}
}

func (cmd ListStacks) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd ListStacks) Run(c *cli.Context) {
	cmd.ui.Say(T("Getting stacks in org {{.OrganizationName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{"OrganizationName": terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	stacks, apiErr := cmd.stacksRepo.FindAll()
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table := terminal.NewTable(cmd.ui, []string{T("name"), T("description")})

	for _, stack := range stacks {
		table.Add(stack.Name, stack.Description)
	}

	table.Print()
}
