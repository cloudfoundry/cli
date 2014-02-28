package service

import (
	"cf"
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
	config                          configuration.Reader
	userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository
	serviceInstanceReq              requirements.ServiceInstanceRequirement
}

func NewUpdateUserProvidedService(ui terminal.UI, config configuration.Reader, userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository) (cmd *UpdateUserProvidedService) {
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

	drainUrl := c.String("l")
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
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	serviceInstance.Params = paramsMap
	serviceInstance.SysLogDrainUrl = drainUrl

	apiResponse := cmd.userProvidedServiceInstanceRepo.Update(serviceInstance.ServiceInstanceFields)
	if apiResponse != nil {
		cmd.ui.Failed(apiResponse.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("TIP: To make these changes take effect, use '%s unbind-service' to unbind the service, '%s bind-service' to rebind, and then '%s push' to update the app with the new env variables", cf.Name(), cf.Name(), cf.Name())

	if params == "" && drainUrl == "" {
		cmd.ui.Warn("No flags specified. No changes were made.")
	}
}
