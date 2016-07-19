package securitygroup

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"

	"code.cloudfoundry.org/cli/cf/api/securitygroups"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type SecurityGroups struct {
	ui                terminal.UI
	securityGroupRepo securitygroups.SecurityGroupRepo
	configRepo        coreconfig.Reader
}

func init() {
	commandregistry.Register(&SecurityGroups{})
}

func (cmd *SecurityGroups) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "security-groups",
		Description: T("List all security groups"),
		Usage: []string{
			"CF_NAME security-groups",
		},
	}
}

func (cmd *SecurityGroups) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd *SecurityGroups) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	return cmd
}

func (cmd *SecurityGroups) Execute(c flags.FlagContext) error {
	cmd.ui.Say(T("Getting security groups as {{.username}}",
		map[string]interface{}{
			"username": terminal.EntityNameColor(cmd.configRepo.Username()),
		}))

	securityGroups, err := cmd.securityGroupRepo.FindAll()
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(securityGroups) == 0 {
		cmd.ui.Say(T("No security groups"))
		return nil
	}

	table := cmd.ui.Table([]string{"", T("Name"), T("Organization"), T("Space")})

	for index, securityGroup := range securityGroups {
		if len(securityGroup.Spaces) > 0 {
			cmd.printSpaces(table, securityGroup, index)
		} else {
			table.Add(fmt.Sprintf("#%d", index), securityGroup.Name, "", "")
		}
	}
	err = table.Print()
	if err != nil {
		return err
	}
	return nil
}

type table interface {
	Add(row ...string)
	Print() error
}

func (cmd SecurityGroups) printSpaces(table table, securityGroup models.SecurityGroup, index int) {
	outputtedIndex := false

	for _, space := range securityGroup.Spaces {
		if !outputtedIndex {
			table.Add(fmt.Sprintf("#%d", index), securityGroup.Name, space.Organization.Name, space.Name)
			outputtedIndex = true
		} else {
			table.Add("", securityGroup.Name, space.Organization.Name, space.Name)
		}
	}
}
