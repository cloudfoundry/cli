package securitygroup

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"

	"code.cloudfoundry.org/cli/cf/api/securitygroups"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type ShowSecurityGroup struct {
	ui                terminal.UI
	securityGroupRepo securitygroups.SecurityGroupRepo
	configRepo        coreconfig.Reader
}

func init() {
	commandregistry.Register(&ShowSecurityGroup{})
}

func (cmd *ShowSecurityGroup) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "security-group",
		Description: T("Show a single security group"),
		Usage: []string{
			T("CF_NAME security-group SECURITY_GROUP"),
		},
	}
}

func (cmd *ShowSecurityGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("security-group"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs, nil
}

func (cmd *ShowSecurityGroup) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	return cmd
}

func (cmd *ShowSecurityGroup) Execute(c flags.FlagContext) error {
	name := c.Args()[0]

	cmd.ui.Say(T("Getting info for security group {{.security_group}} as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(name),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))

	securityGroup, err := cmd.securityGroupRepo.Read(name)
	if err != nil {
		return err
	}

	jsonEncodedBytes, err := json.MarshalIndent(securityGroup.Rules, "\t", "\t")
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	table := cmd.ui.Table([]string{"", ""})
	table.Add(T("Name"), securityGroup.Name)
	table.Add(T("Rules"), "")
	err = table.Print()
	if err != nil {
		return err
	}
	cmd.ui.Say("\t" + string(jsonEncodedBytes))

	cmd.ui.Say("")

	if len(securityGroup.Spaces) > 0 {
		table = cmd.ui.Table([]string{"", T("Organization"), T("Space")})

		for index, space := range securityGroup.Spaces {
			table.Add(fmt.Sprintf("#%d", index), space.Organization.Name, space.Name)
		}
		err = table.Print()
		if err != nil {
			return err
		}
	} else {
		cmd.ui.Say(T("No spaces assigned"))
	}
	return nil
}
