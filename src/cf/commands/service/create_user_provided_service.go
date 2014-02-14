package service

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"encoding/json"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
)

type CreateUserProvidedService struct {
	ui                              terminal.UI
	config                          configuration.Reader
	userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository
}

func NewCreateUserProvidedService(ui terminal.UI, config configuration.Reader, userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository) (cmd CreateUserProvidedService) {
	cmd.ui = ui
	cmd.config = config
	cmd.userProvidedServiceInstanceRepo = userProvidedServiceInstanceRepo
	return
}

func (cmd CreateUserProvidedService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-user-provided-service")
		return
	}

	reqs = []requirements.Requirement{reqFactory.NewLoginRequirement()}
	return
}

func (cmd CreateUserProvidedService) Run(c *cli.Context) {
	name := c.Args()[0]
	drainUrl := c.String("l")

	params := c.String("p")
	params = strings.Trim(params, `"`)
	paramsMap := make(map[string]string)

	err := json.Unmarshal([]byte(params), &paramsMap)
	if err != nil && params != "" {
		paramsMap = cmd.mapValuesFromPrompt(params, paramsMap)
	}

	cmd.ui.Say("Creating user provided service %s in org %s / space %s as %s...",
		terminal.EntityNameColor(name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiResponse := cmd.userProvidedServiceInstanceRepo.Create(name, drainUrl, paramsMap)
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
