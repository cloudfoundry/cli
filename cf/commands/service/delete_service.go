package service

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteService struct {
	ui                 terminal.UI
	config             configuration.Reader
	serviceRepo        api.ServiceRepository
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewDeleteService(ui terminal.UI, config configuration.Reader, serviceRepo api.ServiceRepository) (cmd *DeleteService) {
	cmd = new(DeleteService)
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = serviceRepo
	return
}

func (cmd *DeleteService) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-service",
		ShortName:   "ds",
		Description: T("Delete a service instance"),
		Usage:       T("CF_NAME delete-service SERVICE_INSTANCE [-f]"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")},
		},
	}
}

func (cmd *DeleteService) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	var serviceName string

	if len(c.Args()) == 1 {
		serviceName = c.Args()[0]
	}

	if serviceName == "" {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return
}

func (cmd *DeleteService) Run(c *cli.Context) {
	serviceName := c.Args()[0]

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete(T("service"), serviceName) {
			return
		}
	}

	cmd.ui.Say(T("Deleting service {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceName": terminal.EntityNameColor(serviceName),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	instance, apiErr := cmd.serviceRepo.FindInstanceByName(serviceName)

	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Service {{.ServiceName}} does not exist.", map[string]interface{}{"ServiceName": serviceName}))
		return
	default:
		cmd.ui.Failed(apiErr.Error())
		return
	}

	apiErr = cmd.serviceRepo.DeleteService(instance)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
