package securitygroup

import (
	"encoding/json"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateSecurityGroup struct {
	ui                   terminal.UI
	appSecurityGroupRepo api.SecurityGroupRepo
	spaceRepo            api.SpaceRepository
	configRepo           configuration.Reader
}

func NewCreateSecurityGroup(ui terminal.UI, configRepo configuration.Reader, appSecurityGroupRepo api.SecurityGroupRepo, spaceRepo api.SpaceRepository) CreateSecurityGroup {
	return CreateSecurityGroup{
		ui:                   ui,
		configRepo:           configRepo,
		appSecurityGroupRepo: appSecurityGroupRepo,
		spaceRepo:            spaceRepo,
	}
}

func (cmd CreateSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-security-group",
		Description: "<<< description goes here>>>",
		Usage:       "CF_NAME create-security-group NAME",
		Flags: []cli.Flag{
			flag_helpers.NewStringSliceFlag("rules", "Create Rules Everything Around Me"),
			flag_helpers.NewStringSliceFlag("space", "BOOM A SPACE IS HERE"),
		},
	}
}

func (cmd CreateSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd CreateSecurityGroup) Run(context *cli.Context) {
	name := context.Args()[0]
	rules := context.StringSlice("rules")
	spaces := context.StringSlice("space")
	spaceGuids := []string{}
	for _, spaceName := range spaces {
		space, err := cmd.spaceRepo.FindByName(spaceName)

		if err != nil {
			cmd.ui.Failed("Could not find space named '%s'", spaceName)
		}

		spaceGuids = append(spaceGuids, space.Guid)
	}

	ruleMaps := []map[string]string{}
	for _, rule := range rules {
		ruleMap := map[string]string{}
		err := json.Unmarshal([]byte(rule), &ruleMap)
		if err != nil {
			cmd.ui.Failed("Incorrect json format: %s", err.Error())
		}

		ruleMaps = append(ruleMaps, ruleMap)
	}

	cmd.ui.Say("Creating application security group '%s' as '%s'", name, cmd.configRepo.Username())

	err := cmd.appSecurityGroupRepo.Create(name, ruleMaps, spaceGuids)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
