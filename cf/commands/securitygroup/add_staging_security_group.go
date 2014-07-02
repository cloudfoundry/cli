package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/api/security_groups/defaults/staging"
	"github.com/cloudfoundry/cli/cf/command"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type addToStagingGroup struct {
	ui                terminal.UI
	configRepo        configuration.Reader
	securityGroupRepo security_groups.SecurityGroupRepo
	stagingGroupRepo  staging.StagingSecurityGroupsRepo
}

func NewAddToStagingGroup(ui terminal.UI, configRepo configuration.Reader, securityGroupRepo security_groups.SecurityGroupRepo, stagingGroupRepo staging.StagingSecurityGroupsRepo) command.Command {
	return &addToStagingGroup{
		ui:                ui,
		configRepo:        configRepo,
		securityGroupRepo: securityGroupRepo,
		stagingGroupRepo:  stagingGroupRepo,
	}
}

func (cmd *addToStagingGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "add-staging-security-group",
		Description: "Add a security group to the list of security groups to be used for staging applications",
		Usage:       "CF_NAME add-staging-security-group SECURITY_GROUP",
	}
}

func (cmd *addToStagingGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd *addToStagingGroup) Run(context *cli.Context) {
	name := context.Args()[0]

	securityGroup, err := cmd.securityGroupRepo.Read(name)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Say("Adding security group %s to staging as %s",
		terminal.EntityNameColor(securityGroup.Name),
		terminal.EntityNameColor(cmd.configRepo.Username()))

	err = cmd.stagingGroupRepo.AddToStagingSet(securityGroup.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
