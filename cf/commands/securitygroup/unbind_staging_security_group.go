package securitygroup

import (
	"fmt"

	"code.cloudfoundry.org/cli/v7/cf/api/securitygroups"
	"code.cloudfoundry.org/cli/v7/cf/api/securitygroups/defaults/staging"
	"code.cloudfoundry.org/cli/v7/cf/commandregistry"
	"code.cloudfoundry.org/cli/v7/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v7/cf/errors"
	"code.cloudfoundry.org/cli/v7/cf/flags"
	. "code.cloudfoundry.org/cli/v7/cf/i18n"
	"code.cloudfoundry.org/cli/v7/cf/requirements"
	"code.cloudfoundry.org/cli/v7/cf/terminal"
)

type unbindFromStagingGroup struct {
	ui                terminal.UI
	configRepo        coreconfig.Reader
	securityGroupRepo securitygroups.SecurityGroupRepo
	stagingGroupRepo  staging.SecurityGroupsRepo
}

func init() {
	commandregistry.Register(&unbindFromStagingGroup{})
}

func (cmd *unbindFromStagingGroup) MetaData() commandregistry.CommandMetadata {
	primaryUsage := T("CF_NAME unbind-staging-security-group SECURITY_GROUP")
	tipUsage := T("TIP: Changes will not apply to existing running applications until they are restarted.")
	return commandregistry.CommandMetadata{
		Name:        "unbind-staging-security-group",
		Description: T("Unbind a security group from the set of security groups for staging applications"),
		Usage: []string{
			primaryUsage,
			"\n\n",
			tipUsage,
		},
	}
}

func (cmd *unbindFromStagingGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("unbind-staging-security-group"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return reqs, nil
}

func (cmd *unbindFromStagingGroup) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	cmd.stagingGroupRepo = deps.RepoLocator.GetStagingSecurityGroupsRepository()
	return cmd
}

func (cmd *unbindFromStagingGroup) Execute(context flags.FlagContext) error {
	name := context.Args()[0]

	cmd.ui.Say(T("Unbinding security group {{.security_group}} from defaults for staging as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(name),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))

	securityGroup, err := cmd.securityGroupRepo.Read(name)
	switch (err).(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Security group {{.security_group}} {{.error_message}}",
			map[string]interface{}{
				"security_group": terminal.EntityNameColor(name),
				"error_message":  terminal.WarningColor(T("does not exist.")),
			}))
		return nil
	default:
		return err
	}

	err = cmd.stagingGroupRepo.UnbindFromStagingSet(securityGroup.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say("\n\n")
	cmd.ui.Say(T("TIP: Changes will not apply to existing running applications until they are restarted."))
	return nil
}
