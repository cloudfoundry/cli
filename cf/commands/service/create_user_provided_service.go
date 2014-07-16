package service

import (
	"encoding/json"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"strings"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
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

func (cmd CreateUserProvidedService) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-user-provided-service",
		ShortName:   "cups",
		Description: T("Make a user-provided service instance available to cf apps"),
		Usage: T(`CF_NAME create-user-provided-service SERVICE_INSTANCE [-p PARAMETERS] [-l SYSLOG-DRAIN-URL]

   Pass comma separated parameter names to enable interactive mode:
   CF_NAME create-user-provided-service SERVICE_INSTANCE -p "comma, separated, parameter, names"

   Pass parameters as JSON to create a service non-interactively:
   CF_NAME create-user-provided-service SERVICE_INSTANCE -p '{"name":"value","name":"value"}'

EXAMPLE:
   CF_NAME create-user-provided-service oracle-db-mine -p "host, port, dbname, username, password"
   CF_NAME create-user-provided-service oracle-db-mine -p '{"username":"admin","password":"pa55woRD"}'
   CF_NAME create-user-provided-service my-drain-service -l syslog://example.com
`),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("p", T("Parameters")),
			flag_helpers.NewStringFlag("l", T("Syslog Drain Url")),
		},
	}
}

func (cmd CreateUserProvidedService) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = append(reqs, requirementsFactory.NewLoginRequirement())
	return
}

func (cmd CreateUserProvidedService) Run(c *cli.Context) {
	name := c.Args()[0]
	drainUrl := c.String("l")

	params := c.String("p")
	params = strings.Trim(params, `"`)
	paramsMap := make(map[string]interface{})

	err := json.Unmarshal([]byte(params), &paramsMap)
	if err != nil && params != "" {
		paramsMap = cmd.mapValuesFromPrompt(params, paramsMap)
	}

	cmd.ui.Say(T("Creating user provided service {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceName": terminal.EntityNameColor(name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	apiErr := cmd.userProvidedServiceInstanceRepo.Create(name, drainUrl, paramsMap)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}

func (cmd CreateUserProvidedService) mapValuesFromPrompt(params string, paramsMap map[string]interface{}) map[string]interface{} {
	for _, param := range strings.Split(params, ",") {
		param = strings.Trim(param, " ")
		paramsMap[param] = cmd.ui.Ask("%s", param)
	}
	return paramsMap
}
