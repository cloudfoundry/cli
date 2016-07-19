package securitygroup

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/securitygroups"
	"code.cloudfoundry.org/cli/cf/api/securitygroups/defaults/staging"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type bindToStagingGroup struct {
	ui                terminal.UI
	configRepo        coreconfig.Reader
	securityGroupRepo securitygroups.SecurityGroupRepo
	stagingGroupRepo  staging.SecurityGroupsRepo
}

func init() {
	commandregistry.Register(&bindToStagingGroup{})
}

func (cmd *bindToStagingGroup) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "bind-staging-security-group",
		Description: T("Bind a security group to the list of security groups to be used for staging applications"),
		Usage: []string{
			T("CF_NAME bind-staging-security-group SECURITY_GROUP"),
		},
	}
}

func (cmd *bindToStagingGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("bind-staging-security-group"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd *bindToStagingGroup) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	cmd.stagingGroupRepo = deps.RepoLocator.GetStagingSecurityGroupsRepository()
	return cmd
}

func (cmd *bindToStagingGroup) Execute(context flags.FlagContext) error {
	name := context.Args()[0]

	securityGroup, err := cmd.securityGroupRepo.Read(name)
	if err != nil {
		return err
	}

	cmd.ui.Say(T("Binding security group {{.security_group}} to staging as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(securityGroup.Name),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))

	err = cmd.stagingGroupRepo.BindToStagingSet(securityGroup.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
