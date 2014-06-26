package securitygroup

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type SecurityGroups struct {
	ui                terminal.UI
	securityGroupRepo api.SecurityGroupRepo
	configRepo        configuration.Reader
}

func NewSecurityGroups(ui terminal.UI, configRepo configuration.Reader, securityGroupRepo api.SecurityGroupRepo) SecurityGroups {
	return SecurityGroups{
		ui:                ui,
		configRepo:        configRepo,
		securityGroupRepo: securityGroupRepo,
	}
}

func (cmd SecurityGroups) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "security-groups",
		Description: "List all security groups",
		Usage:       "CF_NAME security-group",
	}
}

func (cmd SecurityGroups) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 0 {
		cmd.ui.FailWithUsage(context)
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd SecurityGroups) Run(context *cli.Context) {
	cmd.ui.Say("Getting security groups as '%s'", cmd.configRepo.Username())

	securityGroups, err := cmd.securityGroupRepo.FindAll()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if len(securityGroups) == 0 {
		cmd.ui.Say("No security groups")
		return
	}

	table := terminal.NewTable(cmd.ui, []string{"", "Name", "Organization", "Space"})

	for index, securityGroup := range securityGroups {
		if len(securityGroup.Spaces) > 0 {
			cmd.printSpaces(table, securityGroup, index)
		} else {
			table.Add(fmt.Sprintf("#%d", index), securityGroup.Name, "", "")
		}
	}
	table.Print()
}

func (cmd SecurityGroups) printSpaces(table terminal.Table, securityGroup models.SecurityGroup, index int) {
	outputted_index := false

	for _, space := range securityGroup.Spaces {
		if !outputted_index {
			table.Add(fmt.Sprintf("#%d", index), securityGroup.Name, space.Organization.Name, space.Name)
			outputted_index = true
		} else {
			table.Add("", securityGroup.Name, space.Organization.Name, space.Name)
		}
	}
}
