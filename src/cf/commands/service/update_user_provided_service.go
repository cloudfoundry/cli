package service

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"encoding/json"
	"errors"
	"github.com/codegangsta/cli"
)

type UpdateUserProvidedService struct {
	ui                              terminal.UI
	config                          *configuration.Configuration
	userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository
	serviceInstanceReq              requirements.ServiceInstanceRequirement
}

func NewUpdateUserProvidedService(ui terminal.UI, config *configuration.Configuration, userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository) (cmd *UpdateUserProvidedService) {
	cmd = new(UpdateUserProvidedService)
	cmd.ui = ui
	cmd.config = config
	cmd.userProvidedServiceInstanceRepo = userProvidedServiceInstanceRepo
	return
}

func (cmd *UpdateUserProvidedService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
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

	params := c.String("p")
	paramsMap := make(map[string]string)
	if params != "" {

		err := json.Unmarshal([]byte(params), &paramsMap)
		if err != nil {
			cmd.ui.Failed("JSON is invalid: %s", err.Error())
			return
		}
	}

	cmd.ui.Say("Updating user provided service %s in org %s / space %s as %s...",
		terminal.EntityNameColor(serviceInstance.Name),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	serviceInstance.Params = paramsMap
	serviceInstance.SysLogDrainUrl = c.String("l")

	apiResponse := cmd.userProvidedServiceInstanceRepo.Update(serviceInstance)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
