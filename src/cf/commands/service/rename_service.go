package service

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type RenameService struct {
	ui                 terminal.UI
	serviceRepo        api.ServiceRepository
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewRenameService(ui terminal.UI, serviceRepo api.ServiceRepository) (cmd *RenameService) {
	cmd = new(RenameService)
	cmd.ui = ui
	cmd.serviceRepo = serviceRepo
	return
}

func (cmd *RenameService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("incorrect usage")
		cmd.ui.FailWithUsage(c, "rename-service")
		return
	}

	cmd.serviceInstanceReq = reqFactory.NewServiceInstanceRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
		cmd.serviceInstanceReq,
	}

	return
}

func (cmd *RenameService) Run(c *cli.Context) {
	newName := c.Args()[1]
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()

	cmd.ui.Say("Renaming service %s...", serviceInstance.Name)
	apiResponse := cmd.serviceRepo.RenameService(serviceInstance, newName)
	if apiResponse.IsNotSuccessful() {
		if apiResponse.ErrorCode == cf.SERVICE_INSTANCE_NAME_TAKEN {
			cmd.ui.Failed("%s\nTIP: Use '%s services' to view all services in this org and space.", apiResponse.Message, cf.Name)
		} else {
			cmd.ui.Failed(apiResponse.Message)
		}
		return
	}
	cmd.ui.Ok()
}
