package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/json"
	"github.com/codegangsta/cli"
)

type UpdateSecurityGroup struct {
	ui                terminal.UI
	securityGroupRepo security_groups.SecurityGroupRepo
	configRepo        configuration.Reader
}

func NewUpdateSecurityGroup(ui terminal.UI, configRepo configuration.Reader, securityGroupRepo security_groups.SecurityGroupRepo) UpdateSecurityGroup {
	return UpdateSecurityGroup{
		ui:                ui,
		configRepo:        configRepo,
		securityGroupRepo: securityGroupRepo,
	}
}

func (cmd UpdateSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "update-security-group",
		Description: "Update a security group",
		Usage:       "CF_NAME update-security-group SECURITY_GROUP [--json PATH_TO_JSON_FILE]",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("json", "Path to a file containing rules in JSON format"),
		},
	}
}

func (cmd UpdateSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd UpdateSecurityGroup) Run(context *cli.Context) {
	name := context.Args()[0]
	securityGroup, err := cmd.securityGroupRepo.Read(name)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	pathToJSONFile := context.String("json")
	rules, err := json.ParseJSON(pathToJSONFile)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Say("Updating security group %s as %s",
		terminal.EntityNameColor(name),
		terminal.EntityNameColor(cmd.configRepo.Username()))
	err = cmd.securityGroupRepo.Update(securityGroup.Guid, rules)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
