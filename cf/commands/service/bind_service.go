package service

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type BindService struct {
	ui                 terminal.UI
	config             configuration.Reader
	serviceBindingRepo api.ServiceBindingRepository
	appReq             requirements.ApplicationRequirement
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

type ServiceBinder interface {
	BindApplication(app models.Application, serviceInstance models.ServiceInstance) (apiErr error)
}

func NewBindService(ui terminal.UI, config configuration.Reader, serviceBindingRepo api.ServiceBindingRepository) (cmd *BindService) {
	cmd = new(BindService)
	cmd.ui = ui
	cmd.config = config
	cmd.serviceBindingRepo = serviceBindingRepo
	return
}

func (cmd *BindService) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "bind-service",
		ShortName:   "bs",
		Description: "Bind a service instance to an app",
		Usage:       "CF_NAME bind-service APP SERVICE_INSTANCE",
	}
}

func (cmd *BindService) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {

	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "bind-service")
		return
	}
	appName := c.Args()[0]
	serviceName := c.Args()[1]

	cmd.appReq = requirementsFactory.NewApplicationRequirement(appName)
	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(serviceName)

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement(), cmd.appReq, cmd.serviceInstanceReq}
	return
}

func (cmd *BindService) Run(c *cli.Context) {
	app := cmd.appReq.GetApplication()
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()

	cmd.ui.Say("Binding service %s to app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(serviceInstance.Name),
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	err := cmd.BindApplication(app, serviceInstance)
	if err != nil {
		if err, ok := err.(errors.HttpError); ok && err.ErrorCode() == errors.APP_ALREADY_BOUND {
			cmd.ui.Ok()
			cmd.ui.Warn("App %s is already bound to %s.", app.Name, serviceInstance.Name)
			return
		} else {
			cmd.ui.Failed(err.Error())
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say("TIP: Use '%s' to ensure your env variable changes take effect", terminal.CommandColor(cf.Name()+" push"))
}

func (cmd *BindService) BindApplication(app models.Application, serviceInstance models.ServiceInstance) (apiErr error) {
	apiErr = cmd.serviceBindingRepo.Create(serviceInstance.Guid, app.Guid)
	return
}
