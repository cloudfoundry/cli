package securitygroup

import (
	"encoding/json"
	"fmt"
	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ShowSecurityGroup struct {
	ui                terminal.UI
	securityGroupRepo security_groups.SecurityGroupRepo
	configRepo        configuration.Reader
}

func NewShowSecurityGroup(ui terminal.UI, configRepo configuration.Reader, securityGroupRepo security_groups.SecurityGroupRepo) ShowSecurityGroup {
	return ShowSecurityGroup{
		ui:                ui,
		configRepo:        configRepo,
		securityGroupRepo: securityGroupRepo,
	}
}

func (cmd ShowSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "security-group",
		Description: T("Show a single security group"),
		Usage:       T("CF_NAME security-group SECURITY_GROUP"),
	}
}

func (cmd ShowSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd ShowSecurityGroup) Run(context *cli.Context) {
	name := context.Args()[0]

	cmd.ui.Say(T("Getting info for security group {{.security_group}} as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(name),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))

	securityGroup, err := cmd.securityGroupRepo.Read(name)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	jsonEncodedBytes, encodingErr := json.MarshalIndent(securityGroup.Rules, "\t", "\t")
	if encodingErr != nil {
		cmd.ui.Failed(encodingErr.Error())
	}

	cmd.ui.Ok()
	table := terminal.NewTable(cmd.ui, []string{"", ""})
	table.Add(T("Name"), securityGroup.Name)
	table.Add(T("Rules"), "")
	table.Print()
	cmd.ui.Say("\t" + string(jsonEncodedBytes))

	cmd.ui.Say("")

	if len(securityGroup.Spaces) > 0 {
		table = terminal.NewTable(cmd.ui, []string{"", T("Organization"), T("Space")})

		for index, space := range securityGroup.Spaces {
			table.Add(fmt.Sprintf("#%d", index), space.Organization.Name, space.Name)
		}
		table.Print()
	} else {
		cmd.ui.Say(T("No spaces assigned"))
	}
}
