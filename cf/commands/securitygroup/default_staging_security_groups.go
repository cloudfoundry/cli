package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/security_groups/defaults/staging"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type listDefaultStagingSecurityGroups struct {
	ui                       terminal.UI
	stagingSecurityGroupRepo staging.StagingSecurityGroupsRepo
	configRepo               configuration.Reader
}

func NewListDefaultStagingSecurityGroups(ui terminal.UI, configRepo configuration.Reader, stagingSecurityGroupRepo staging.StagingSecurityGroupsRepo) listDefaultStagingSecurityGroups {
	return listDefaultStagingSecurityGroups{
		ui:                       ui,
		configRepo:               configRepo,
		stagingSecurityGroupRepo: stagingSecurityGroupRepo,
	}
}

func (cmd listDefaultStagingSecurityGroups) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "default-staging-security-groups",
		Description: "List security groups in the set of default security groups for staging applications.",
		Usage:       "CF_NAME default-security-staging-groups",
	}
}

func (cmd listDefaultStagingSecurityGroups) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 0 {
		cmd.ui.FailWithUsage(context)
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd listDefaultStagingSecurityGroups) Run(context *cli.Context) {
	cmd.ui.Say("Acquiring default security groups as '%s'", terminal.EntityNameColor(cmd.configRepo.Username()))

	defaultSecurityGroupsFields, err := cmd.stagingSecurityGroupRepo.List()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(defaultSecurityGroupsFields) > 0 {
		for _, value := range defaultSecurityGroupsFields {
			cmd.ui.Say(value.Name)
		}
	} else {
		cmd.ui.Say("No default staging security groups set")
	}
}
