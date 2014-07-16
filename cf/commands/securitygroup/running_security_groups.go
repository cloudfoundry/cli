package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/security_groups/defaults/running"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type listRunningSecurityGroups struct {
	ui                       terminal.UI
	runningSecurityGroupRepo running.RunningSecurityGroupsRepo
	configRepo               configuration.Reader
}

func NewListRunningSecurityGroups(ui terminal.UI, configRepo configuration.Reader, runningSecurityGroupRepo running.RunningSecurityGroupsRepo) listRunningSecurityGroups {
	return listRunningSecurityGroups{
		ui:                       ui,
		configRepo:               configRepo,
		runningSecurityGroupRepo: runningSecurityGroupRepo,
	}
}

func (cmd listRunningSecurityGroups) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "running-security-groups",
		Description: T("List security groups in the set of security groups for running applications"),
		Usage:       "CF_NAME running-security-groups",
	}
}

func (cmd listRunningSecurityGroups) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 0 {
		cmd.ui.FailWithUsage(context)
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd listRunningSecurityGroups) Run(context *cli.Context) {
	cmd.ui.Say(T("Acquiring running security groups as '{{.username}}'", map[string]interface{}{
		"username": terminal.EntityNameColor(cmd.configRepo.Username()),
	}))

	defaultSecurityGroupsFields, err := cmd.runningSecurityGroupRepo.List()
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
		cmd.ui.Say(T("No running security groups set"))
	}
}
