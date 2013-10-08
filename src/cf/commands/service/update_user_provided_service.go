package service

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"encoding/json"
	"errors"
	"github.com/codegangsta/cli"
)

type UpdateUserProvidedService struct {
	ui                 terminal.UI
	serviceRepo        api.ServiceRepository
	serviceInstanceReq requirements.ServiceInstanceRequirement
}

func NewUpdateUserProvidedService(ui terminal.UI, sR api.ServiceRepository) (cmd *UpdateUserProvidedService) {
	cmd = new(UpdateUserProvidedService)
	cmd.ui = ui
	cmd.serviceRepo = sR
	return
}

func (cmd *UpdateUserProvidedService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "update-user-provided-service")
		return
	}

	cmd.serviceInstanceReq = reqFactory.NewServiceInstanceRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.serviceInstanceReq,
	}

	return
}

func (cmd *UpdateUserProvidedService) Run(c *cli.Context) {
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()
	if !serviceInstance.IsUserProvided() {
		cmd.ui.Failed("Service Instance is not user provided")
		return
	}

	params := c.Args()[1]
	paramsMap := make(map[string]string)

	err := json.Unmarshal([]byte(params), &paramsMap)
	if err != nil {
		cmd.ui.Failed("JSON is invalid: %s", err.Error())
		return
	}

	cmd.ui.Say("Updating user provided service %s...", serviceInstance.Name)

	apiResponse := cmd.serviceRepo.UpdateUserProvidedServiceInstance(serviceInstance, paramsMap)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
