package securitygroup

import (
	"encoding/json"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateSecurityGroup struct {
	ui                terminal.UI
	securityGroupRepo security_groups.SecurityGroupRepo
	spaceRepo         api.SpaceRepository
	configRepo        configuration.Reader
}

func NewCreateSecurityGroup(ui terminal.UI, configRepo configuration.Reader, securityGroupRepo security_groups.SecurityGroupRepo, spaceRepo api.SpaceRepository) CreateSecurityGroup {
	return CreateSecurityGroup{
		ui:                ui,
		configRepo:        configRepo,
		securityGroupRepo: securityGroupRepo,
		spaceRepo:         spaceRepo,
	}
}

func (cmd CreateSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-security-group",
		Description: "create a security group",
		Usage:       "CF_NAME create-security-group NAME [--rules RULES] [--space SpaceName]",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("rules", "JSON encoded array of rules"),
			flag_helpers.NewStringSliceFlag("space", "The name of a space to apply this rule to. Can be provided multiple times"),
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
	rules := context.String("rules")
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
	if rules != "" {
		err := json.Unmarshal([]byte(rules), &ruleMaps)
		if err != nil {
			cmd.ui.Failed("Incorrect json format: %s", err.Error())
		}
	}

	cmd.ui.Say("Creating security group '%s' as '%s', applying to %d spaces", name, cmd.configRepo.Username(), len(spaceGuids))

	err := cmd.securityGroupRepo.Create(name, ruleMaps, spaceGuids)

	httpErr, ok := err.(errors.HttpError)
	if ok && httpErr.ErrorCode() == errors.SECURITY_GROUP_EXISTS {
		cmd.ui.Ok()
		cmd.ui.Warn("Security group '%s' already exists", name)
		return
	}

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
