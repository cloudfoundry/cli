package securitygroup

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/api/security_groups/defaults/staging"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type unbindFromStagingGroup struct {
	ui                terminal.UI
	configRepo        core_config.Reader
	securityGroupRepo security_groups.SecurityGroupRepo
	stagingGroupRepo  staging.StagingSecurityGroupsRepo
}

func init() {
	command_registry.Register(&unbindFromStagingGroup{})
}

func (cmd *unbindFromStagingGroup) MetaData() command_registry.CommandMetadata {
	primaryUsage := T("CF_NAME unbind-staging-security-group SECURITY_GROUP")
	tipUsage := T("TIP: Changes will not apply to existing running applications until they are restarted.")
	return command_registry.CommandMetadata{
		Name:        "unbind-staging-security-group",
		Description: T("Unbind a security group from the set of security groups for staging applications"),
		Usage:       strings.Join([]string{primaryUsage, tipUsage}, "\n\n"),
	}
}

func (cmd *unbindFromStagingGroup) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("unbind-staging-security-group"))
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd *unbindFromStagingGroup) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.configRepo = deps.Config
	cmd.securityGroupRepo = deps.RepoLocator.GetSecurityGroupRepository()
	cmd.stagingGroupRepo = deps.RepoLocator.GetStagingSecurityGroupsRepository()
	return cmd
}

func (cmd *unbindFromStagingGroup) Execute(context flags.FlagContext) {
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
		return
	default:
		cmd.ui.Failed(err.Error())
	}

	err = cmd.stagingGroupRepo.UnbindFromStagingSet(securityGroup.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("\n\n")
	cmd.ui.Say(T("TIP: Changes will not apply to existing running applications until they are restarted."))
}
