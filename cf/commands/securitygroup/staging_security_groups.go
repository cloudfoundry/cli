package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/security_groups/defaults/staging"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type listStagingSecurityGroups struct {
	ui                       terminal.UI
	stagingSecurityGroupRepo staging.StagingSecurityGroupsRepo
	configRepo               core_config.Reader
}

func init() {
	command_registry.Register(&listStagingSecurityGroups{})
}

func (cmd *listStagingSecurityGroups) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "staging-security-groups",
		Description: T("List security groups in the staging set for applications"),
		Usage:       "CF_NAME staging-security-groups",
	}
}

func (cmd *listStagingSecurityGroups) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("staging-security-groups"))
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd *listStagingSecurityGroups) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.configRepo = deps.Config
	cmd.stagingSecurityGroupRepo = deps.RepoLocator.GetStagingSecurityGroupsRepository()
	return cmd
}

func (cmd *listStagingSecurityGroups) Execute(context flags.FlagContext) {
	cmd.ui.Say(T("Acquiring staging security group as {{.username}}",
		map[string]interface{}{
			"username": terminal.EntityNameColor(cmd.configRepo.Username()),
		}))

	SecurityGroupsFields, err := cmd.stagingSecurityGroupRepo.List()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(SecurityGroupsFields) > 0 {
		for _, value := range SecurityGroupsFields {
			cmd.ui.Say(value.Name)
		}
	} else {
		cmd.ui.Say(T("No staging security group set"))
	}
}
