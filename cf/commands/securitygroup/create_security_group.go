package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/json"
	"github.com/codegangsta/cli"
)

type CreateSecurityGroup struct {
	ui                terminal.UI
	securityGroupRepo security_groups.SecurityGroupRepo
	configRepo        configuration.Reader
}

func NewCreateSecurityGroup(ui terminal.UI, configRepo configuration.Reader, securityGroupRepo security_groups.SecurityGroupRepo) CreateSecurityGroup {
	return CreateSecurityGroup{
		ui:                ui,
		configRepo:        configRepo,
		securityGroupRepo: securityGroupRepo,
	}
}

func (cmd CreateSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-security-group",
		Description: T("Create a security group"),
		Usage:       T("CF_NAME create-security-group SECURITY_GROUP [--json PATH_TO_JSON_FILE]"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("json", T("Path to a file containing rules in JSON format")),
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
	pathToJSONFile := context.String("json")
	rules, err := json.ParseJSON(pathToJSONFile)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Say(T("Creating security group {{.security_group}} as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(name),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))

	err = cmd.securityGroupRepo.Create(name, rules)

	httpErr, ok := err.(errors.HttpError)
	if ok && httpErr.ErrorCode() == errors.SECURITY_GROUP_EXISTS {
		cmd.ui.Ok()
		cmd.ui.Warn(T("Security group {{.security_group}} {{.error_message}}",
			map[string]interface{}{
				"security_group": terminal.EntityNameColor(name),
				"error_message":  terminal.WarningColor(T("already exists")),
			}))
		return
	}

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
