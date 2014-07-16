package securitygroup

import (
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteSecurityGroup struct {
	ui                terminal.UI
	securityGroupRepo security_groups.SecurityGroupRepo
	configRepo        configuration.Reader
}

func NewDeleteSecurityGroup(ui terminal.UI, configRepo configuration.Reader, securityGroupRepo security_groups.SecurityGroupRepo) DeleteSecurityGroup {
	return DeleteSecurityGroup{
		ui:                ui,
		configRepo:        configRepo,
		securityGroupRepo: securityGroupRepo,
	}
}

func (cmd DeleteSecurityGroup) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-security-group",
		Description: T("Deletes a security group"),
		Usage:       T("CF_NAME delete-security-group SECURITY_GROUP [-f]"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")},
		},
	}
}

func (cmd DeleteSecurityGroup) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	requirements := []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return requirements, nil
}

func (cmd DeleteSecurityGroup) Run(context *cli.Context) {
	name := context.Args()[0]
	cmd.ui.Say(T("Deleting security group {{.security_group}} as {{.username}}",
		map[string]interface{}{
			"security_group": terminal.EntityNameColor(name),
			"username":       terminal.EntityNameColor(cmd.configRepo.Username()),
		}))

	if !context.Bool("f") {
		response := cmd.ui.ConfirmDelete(T("security group"), name)
		if !response {
			return
		}
	}

	group, err := cmd.securityGroupRepo.Read(name)
	switch err.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Security group {{.security_group}} does not exist", map[string]interface{}{"security_group": name}))
		return
	default:
		cmd.ui.Failed(err.Error())
	}

	err = cmd.securityGroupRepo.Delete(group.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
