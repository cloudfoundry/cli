package commands

import (
	"github.com/cloudfoundry/cli/cf/api/stacks"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ListStack struct {
	ui         terminal.UI
	config     coreconfig.Reader
	stacksRepo stacks.StackRepository
}

func init() {
	commandregistry.Register(&ListStack{})
}

func (cmd *ListStack) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["guid"] = &flags.BoolFlag{Name: "guid", Usage: T("Retrieve and display the given stack's guid. All other output for the stack is suppressed.")}

	return commandregistry.CommandMetadata{
		Name:        "stack",
		Description: T("Show information for a stack (a stack is a pre-built file system, including an operating system, that can run apps)"),
		Usage: []string{
			T("CF_NAME stack STACK_NAME"),
		},
		Flags:     fs,
		TotalArgs: 1,
	}
}

func (cmd *ListStack) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires stack name as argument\n\n") + commandregistry.Commands.CommandUsage("stack"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *ListStack) SetDependency(deps commandregistry.Dependency, _ bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.stacksRepo = deps.RepoLocator.GetStackRepository()
	return cmd
}

func (cmd *ListStack) Execute(c flags.FlagContext) error {
	stackName := c.Args()[0]

	stack, err := cmd.stacksRepo.FindByName(stackName)

	if c.Bool("guid") {
		cmd.ui.Say(stack.GUID)
	} else {
		if err != nil {
			return err
		}

		cmd.ui.Say(T("Getting stack '{{.Stack}}' in org {{.OrganizationName}} / space {{.SpaceName}} as {{.Username}}...",
			map[string]interface{}{"Stack": stackName,
				"OrganizationName": terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName":        terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"Username":         terminal.EntityNameColor(cmd.config.Username())}))

		cmd.ui.Ok()
		cmd.ui.Say("")
		table := cmd.ui.Table([]string{T("name"), T("description")})
		table.Add(stack.Name, stack.Description)
		table.Print()
	}
	return nil
}
