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

type removeFromDefaultStagingGroup struct {
	ui                terminal.UI
	configRepo        configuration.Reader
	securityGroupRepo security_groups.SecurityGroupRepo
	stagingGroupRepo  staging.StagingSecurityGroupsRepo
}

func NewRemoveFromDefaultStagingGroup(ui terminal.UI, configRepo configuration.Reader, securityGroupRepo security_groups.SecurityGroupRepo, stagingGroupRepo staging.StagingSecurityGroupsRepo) command.Command {
	return &removeFromDefaultStagingGroup{
		ui:                ui,
		configRepo:        configRepo,
		securityGroupRepo: securityGroupRepo,
		stagingGroupRepo:  stagingGroupRepo,
	}
}

func (cmd *removeFromDefaultStagingGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "remove-default-staging-security-group",
		Description: "Remove a security group from the set of default security groups for staging",
		Usage:       "CF_NAME remove-default-staging-security-group NAME",
	}
}

func (cmd *removeFromDefaultStagingGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd *removeFromDefaultStagingGroup) Run(context *cli.Context) {
	name := context.Args()[0]

	securityGroup, err := cmd.securityGroupRepo.Read(name)
	switch (err).(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn("Security group %s %s",
			terminal.EntityNameColor(name),
			terminal.WarningColor("does not exist."))
		return
	default:
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Say("Removing security group %s from defaults for staging as %s",
		terminal.EntityNameColor(securityGroup.Name),
		terminal.EntityNameColor(cmd.configRepo.Username()))
	err = cmd.stagingGroupRepo.RemoveFromDefaultStagingSet(securityGroup.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
