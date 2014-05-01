package service

import (
	"encoding/json"
	"errors"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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

func (command CreateUserProvidedService) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-user-provided-service",
		ShortName:   "cups",
		Description: "Make a user-provided service instance available to cf apps",
		Usage: "CF_NAME create-user-provided-service SERVICE_INSTANCE [-p PARAMETERS] [-l SYSLOG-DRAIN-URL]\n" +
			"\n   Pass comma separated parameter names to enable interactive mode:\n" +
			"   CF_NAME create-user-provided-service SERVICE_INSTANCE -p \"comma, separated, parameter, names\"\n" +
			"\n   Pass parameters as JSON to create a service non-interactively:\n" +
			"   CF_NAME create-user-provided-service SERVICE_INSTANCE -p '{\"name\":\"value\",\"name\":\"value\"}'\n" +
			"\nEXAMPLE:\n" +
			"   CF_NAME create-user-provided-service oracle-db-mine -p \"host, port, dbname, username, password\"\n" +
			"   CF_NAME create-user-provided-service oracle-db-mine -p '{\"username\":\"admin\",\"password\":\"pa55woRD\"}'\n" +
			"   CF_NAME create-user-provided-service my-drain-service -l syslog://example.com\n",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("p", "Parameters"),
			flag_helpers.NewStringFlag("l", "Syslog Drain Url"),
		},
	}
}

func (cmd CreateUserProvidedService) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-user-provided-service")
		return
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
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

	apiErr := cmd.userProvidedServiceInstanceRepo.Create(name, drainUrl, paramsMap)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}

func (cmd CreateUserProvidedService) mapValuesFromPrompt(params string, paramsMap map[string]string) map[string]string {
	for _, param := range strings.Split(params, ",") {
		param = strings.Trim(param, " ")
		paramsMap[param] = cmd.ui.Ask("%s", param)
	}
	return paramsMap
}
