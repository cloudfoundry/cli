package service

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"encoding/json"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
)

type CreateUserProvidedService struct {
	ui                              terminal.UI
	userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository
}

func NewCreateUserProvidedService(ui terminal.UI, userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository) (cmd CreateUserProvidedService) {
	cmd.ui = ui
	cmd.userProvidedServiceInstanceRepo = userProvidedServiceInstanceRepo
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
	params = strings.Trim(params, `"`)
	paramsMap := make(map[string]string)

	err := json.Unmarshal([]byte(params), &paramsMap)
	if err != nil {
		paramsMap = cmd.mapValuesFromPrompt(params, paramsMap)
	}

	cmd.ui.Say("Creating user provided service...")

	apiResponse := cmd.userProvidedServiceInstanceRepo.Create(name, paramsMap)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}

func (cmd CreateUserProvidedService) mapValuesFromPrompt(params string, paramsMap map[string]string) map[string]string {
	for _, param := range strings.Split(params, ",") {
		param = strings.Trim(param, " ")
		paramsMap[param] = cmd.ui.Ask("%s%s", param, terminal.PromptColor(">"))
	}
	return paramsMap
}
