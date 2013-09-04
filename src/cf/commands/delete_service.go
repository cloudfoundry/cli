package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteService struct {
	ui                 term.UI
	config             *configuration.Configuration
	serviceRepo        api.ServiceRepository
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewDeleteService(ui term.UI, config *configuration.Configuration, sR api.ServiceRepository) (cmd *DeleteService) {
	cmd = new(DeleteService)
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = sR
	return
}

func (cmd *DeleteService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []Requirement, err error) {
	cmd.serviceInstanceReq = reqFactory.NewServiceInstanceRequirement(c.Args()[0])

	reqs = []Requirement{&cmd.serviceInstanceReq}
	return
}

func (cmd *DeleteService) Run(c *cli.Context) {
	instance := cmd.serviceInstanceReq.ServiceInstance
	cmd.ui.Say("Deleting service %s...", term.Cyan(instance.Name))

	err := cmd.serviceRepo.DeleteService(cmd.config, instance)
	if err != nil {
		cmd.ui.Failed("Failed deleting service", err)
		return
	}

	cmd.ui.Ok()
}
