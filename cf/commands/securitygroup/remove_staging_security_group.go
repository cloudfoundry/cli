package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/api/security_groups/defaults/staging"
	"github.com/cloudfoundry/cli/cf/command"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type removeFromStagingGroup struct {
	ui                terminal.UI
	configRepo        configuration.Reader
	securityGroupRepo security_groups.SecurityGroupRepo
	stagingGroupRepo  staging.StagingSecurityGroupsRepo
}

func NewRemoveFromStagingGroup(ui terminal.UI, configRepo configuration.Reader, securityGroupRepo security_groups.SecurityGroupRepo, stagingGroupRepo staging.StagingSecurityGroupsRepo) command.Command {
	return &removeFromStagingGroup{
		ui:                ui,
		configRepo:        configRepo,
		securityGroupRepo: securityGroupRepo,
		stagingGroupRepo:  stagingGroupRepo,
	}
}

func (cmd *removeFromStagingGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "remove-staging-security-group",
		Description: T("Remove a security group from the set of default security groups for staging applications"),
		Usage:       T("CF_NAME remove-staging-security-group SECURITY_GROUP"),
	}
}

func (cmd *removeFromStagingGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd *removeFromStagingGroup) Run(context *cli.Context) {
	name := context.Args()[0]

	cmd.ui.Say(T("Removing security group {{.security_group}} from defaults for staging as {{.username}}",
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

	err = cmd.stagingGroupRepo.RemoveFromStagingSet(securityGroup.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
