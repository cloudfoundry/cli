package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/api/security_groups/defaults/running"
	"github.com/cloudfoundry/cli/cf/command"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type removeFromRunningGroup struct {
	ui                terminal.UI
	configRepo        configuration.Reader
	securityGroupRepo security_groups.SecurityGroupRepo
	runningGroupRepo  running.RunningSecurityGroupsRepo
}

func NewRemoveFromRunningGroup(ui terminal.UI, configRepo configuration.Reader, securityGroupRepo security_groups.SecurityGroupRepo, runningGroupRepo running.RunningSecurityGroupsRepo) command.Command {
	return &removeFromRunningGroup{
		ui:                ui,
		configRepo:        configRepo,
		securityGroupRepo: securityGroupRepo,
		runningGroupRepo:  runningGroupRepo,
	}
}

func (cmd *removeFromRunningGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "remove-running-security-group",
		Description: T("Remove a security group from the set of security groups for running applications"),
		Usage:       T("CF_NAME remove-running-security-group SECURITY_GROUP"),
	}
}

func (cmd *removeFromRunningGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}, nil
}

func (cmd *removeFromRunningGroup) Run(context *cli.Context) {
	name := context.Args()[0]

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

	cmd.ui.Say(T("Removing security group {{.security_group}} from defaults for running as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(securityGroup.Name),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))
	err = cmd.runningGroupRepo.RemoveFromRunningSet(securityGroup.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
