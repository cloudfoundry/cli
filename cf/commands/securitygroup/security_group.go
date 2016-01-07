package securitygroup

import (
	"encoding/json"
	"fmt"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ShowSecurityGroup struct {
	ui                terminal.UI
	securityGroupRepo security_groups.SecurityGroupRepo
	configRepo        core_config.Reader
}

func init() {
	command_registry.Register(&ShowSecurityGroup{})
}

func (cmd *ShowSecurityGroup) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "security-group",
		Description: T("Show a single security group"),
		Usage:       T("CF_NAME security-group SECURITY_GROUP"),
	}
}

func (cmd *ShowSecurityGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("security-group"))
	}

	return []requirements.Requirement{requirementsFactory.NewLoginRequirement()}, nil
}

func (cmd *ShowSecurityGroup) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	return cmd
}

func (cmd *ShowSecurityGroup) Execute(c flags.FlagContext) {
	name := c.Args()[0]

	cmd.ui.Say(T("Getting info for security group {{.security_group}} as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(name),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))

	securityGroup, err := cmd.securityGroupRepo.Read(name)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	jsonEncodedBytes, encodingErr := json.MarshalIndent(securityGroup.Rules, "\t", "\t")
	if encodingErr != nil {
		cmd.ui.Failed(encodingErr.Error())
	}

	cmd.ui.Ok()
	table := terminal.NewTable(cmd.ui, []string{"", ""})
	table.Add(T("Name"), securityGroup.Name)
	table.Add(T("Rules"), "")
	table.Print()
	cmd.ui.Say("\t" + string(jsonEncodedBytes))

	cmd.ui.Say("")

	if len(securityGroup.Spaces) > 0 {
		table = terminal.NewTable(cmd.ui, []string{"", T("Organization"), T("Space")})

		for index, space := range securityGroup.Spaces {
			table.Add(fmt.Sprintf("#%d", index), space.Organization.Name, space.Name)
		}
		table.Print()
	} else {
		cmd.ui.Say(T("No spaces assigned"))
	}
}
