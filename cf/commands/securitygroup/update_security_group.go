package securitygroup

import (
	. "github.com/cloudfoundry/cli/cf/i18n"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
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
	primaryUsage := T("CF_NAME update-security-group SECURITY_GROUP PATH_TO_JSON_RULES_FILE")
	secondaryUsage := T("   The provided path can be an absolute or relative path to a file.\n   It should have a single array with JSON objects inside describing the rules.")
	return command_metadata.CommandMetadata{
		Name:        "update-security-group",
		Description: T("Update a security group"),
		Usage:       strings.Join([]string{primaryUsage, secondaryUsage}, "\n\n"),
	}
}

func (cmd UpdateSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 2 {
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

	pathToJSONFile := context.Args()[1]
	rules, err := json.ParseJSON(pathToJSONFile)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Say(T("Updating security group {{.security_group}} as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(name),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))
	err = cmd.securityGroupRepo.Update(securityGroup.Guid, rules)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
