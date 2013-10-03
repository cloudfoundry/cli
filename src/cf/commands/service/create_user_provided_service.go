package service

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
)

type CreateUserProvidedService struct {
	ui          terminal.UI
	serviceRepo api.ServiceRepository
}

func NewCreateUserProvidedService(ui terminal.UI, sR api.ServiceRepository) (cmd CreateUserProvidedService) {
	cmd.ui = ui
	cmd.serviceRepo = sR
	return
}

func (cmd CreateUserProvidedService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-user-provided-service")
		return
	}

	return
}

func (cmd CreateUserProvidedService) Run(c *cli.Context) {
	name := c.Args()[0]
	params := c.Args()[1]
	paramsMap := make(map[string]string)
	params = strings.Trim(params, `"`)

	println("PARAMS", params)

	for _, param := range strings.Split(params, ",") {
		param = strings.Trim(param, " ")
		paramsMap[param] = cmd.ui.Ask("%s%s", param, terminal.PromptColor(">"))
	}

	cmd.ui.Say("Creating service...")

	apiStatus := cmd.serviceRepo.CreateUserProvidedServiceInstance(name, paramsMap)
	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	cmd.ui.Ok()
}
