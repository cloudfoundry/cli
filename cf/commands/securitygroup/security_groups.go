package securitygroup

import (
	"fmt"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type SecurityGroups struct {
	ui                terminal.UI
	securityGroupRepo security_groups.SecurityGroupRepo
	configRepo        core_config.Reader
}

func init() {
	command_registry.Register(&SecurityGroups{})
}

func (cmd *SecurityGroups) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "security-groups",
		Description: T("List all security groups"),
		Usage:       "CF_NAME security-groups",
	}
}

func (cmd *SecurityGroups) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("security-groups"))
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd *SecurityGroups) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	return cmd
}

func (cmd *SecurityGroups) Execute(c flags.FlagContext) {
	cmd.ui.Say(T("Getting security groups as {{.username}}",
		map[string]interface{}{
			"username": terminal.EntityNameColor(cmd.configRepo.Username()),
		}))

	securityGroups, err := cmd.securityGroupRepo.FindAll()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(securityGroups) == 0 {
		cmd.ui.Say(T("No security groups"))
		return
	}

	table := terminal.NewTable(cmd.ui, []string{"", T("Name"), T("Organization"), T("Space")})

	for index, securityGroup := range securityGroups {
		if len(securityGroup.Spaces) > 0 {
			cmd.printSpaces(table, securityGroup, index)
		} else {
			table.Add(fmt.Sprintf("#%d", index), securityGroup.Name, "", "")
		}
	}
	table.Print()
}

func (cmd SecurityGroups) printSpaces(table terminal.Table, securityGroup models.SecurityGroup, index int) {
	outputted_index := false

	for _, space := range securityGroup.Spaces {
		if !outputted_index {
			table.Add(fmt.Sprintf("#%d", index), securityGroup.Name, space.Organization.Name, space.Name)
			outputted_index = true
		} else {
			table.Add("", securityGroup.Name, space.Organization.Name, space.Name)
		}
	}
}
