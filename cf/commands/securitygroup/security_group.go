package securitygroup

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ShowSecurityGroup struct {
	ui                   terminal.UI
	appSecurityGroupRepo security_groups.SecurityGroupRepo
	configRepo           configuration.Reader
}

func NewShowAppSecurityGroup(ui terminal.UI, configRepo configuration.Reader, appSecurityGroupRepo security_groups.SecurityGroupRepo) ShowSecurityGroup {
	return ShowSecurityGroup{
		ui:                   ui,
		configRepo:           configRepo,
		appSecurityGroupRepo: appSecurityGroupRepo,
	}
}

func (cmd ShowSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "security-group",
		Description: "Show a single security group",
		Usage:       "CF_NAME security-group NAME",
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

	cmd.ui.Say("Getting info for application security group '%s' as '%s'", name, cmd.configRepo.Username())

	appSecurityGroup, err := cmd.appSecurityGroupRepo.Read(name)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	jsonEncodedBytes, encodingErr := json.Marshal(appSecurityGroup.Rules)
	if encodingErr != nil {
		cmd.ui.Failed(encodingErr.Error())
	}

	cmd.ui.Ok()
	table := terminal.NewTable(cmd.ui, []string{"", ""})
	table.Add("Name", appSecurityGroup.Name)
	table.Add("Rules", string(jsonEncodedBytes))
	table.Print()
	cmd.ui.Say("")

	if len(appSecurityGroup.Spaces) > 0 {
		table = terminal.NewTable(cmd.ui, []string{"", "Organization", "Space"})

		for index, space := range appSecurityGroup.Spaces {
			table.Add(fmt.Sprintf("#%d", index), space.Organization.Name, space.Name)
		}
		table.Print()
	} else {
		cmd.ui.Say("No spaces assigned")
	}
}
