package appsecuritygroup

import (
	"encoding/json"
	"strings"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ShowApplicationSecurityGroup struct {
	ui                   terminal.UI
	appSecurityGroupRepo api.SecurityGroupRepo
	configRepo           configuration.Reader
}

func NewShowAppSecurityGroup(ui terminal.UI, configRepo configuration.Reader, appSecurityGroupRepo api.SecurityGroupRepo) ShowApplicationSecurityGroup {
	return ShowApplicationSecurityGroup{
		ui:                   ui,
		configRepo:           configRepo,
		appSecurityGroupRepo: appSecurityGroupRepo,
	}
}

func (cmd ShowApplicationSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "application-security-group",
		Description: "<<< Description goes here >>>",
		Usage:       "CF_NAME application-security-group NAME",
	}
}

func (cmd ShowApplicationSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd ShowApplicationSecurityGroup) Run(context *cli.Context) {
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

	spaceNames := []string{}
	for _, space := range appSecurityGroup.Spaces {
		spaceNames = append(spaceNames, space.Name)
	}

	cmd.ui.Ok()
	table := terminal.NewTable(cmd.ui, []string{"", ""})
	table.Add("Name:", appSecurityGroup.Name)
	table.Add("Rules:", string(jsonEncodedBytes))

	if len(spaceNames) == 0 {
		table.Add("Spaces:", "No spaces")
	} else {
		table.Add("Spaces:", strings.Join(spaceNames, ", "))
	}

	table.Print()
}
