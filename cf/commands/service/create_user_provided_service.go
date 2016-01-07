package service

import (
	"encoding/json"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type CreateUserProvidedService struct {
	ui                              terminal.UI
	config                          core_config.Reader
	userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository
}

func init() {
	command_registry.Register(&CreateUserProvidedService{})
}

func (cmd *CreateUserProvidedService) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["p"] = &cliFlags.StringFlag{ShortName: "p", Usage: T("Credentials")}
	fs["l"] = &cliFlags.StringFlag{ShortName: "l", Usage: T("Syslog Drain Url")}

	return command_registry.CommandMetadata{
		Name:        "create-user-provided-service",
		ShortName:   "cups",
		Description: T("Make a user-provided service instance available to cf apps"),
		Usage: T(`CF_NAME create-user-provided-service SERVICE_INSTANCE [-p CREDENTIALS] [-l SYSLOG-DRAIN-URL]

   Pass comma separated credential parameter names to enable interactive mode:
   CF_NAME create-user-provided-service SERVICE_INSTANCE -p "comma, separated, parameter, names"

   Pass credential parameters as JSON to create a service non-interactively:
   CF_NAME create-user-provided-service SERVICE_INSTANCE -p '{"name":"value","name":"value"}'

EXAMPLE 
      CF_NAME create-user-provided-service my-db-mine -p "username, password"
      CF_NAME create-user-provided-service my-drain-service -l syslog://example.com

   Linux/Mac:
      CF_NAME create-user-provided-service my-db-mine -p '{"username":"admin","password":"pa55woRD"}'

   Windows Command Line
      CF_NAME create-user-provided-service my-db-mine -p "{\"username\":\"admin\",\"password\":\"pa55woRD\"}"

   Windows PowerShell
      CF_NAME create-user-provided-service my-db-mine -p '{\"username\":\"admin\",\"password\":\"pa55woRD\"}'
`),
		Flags: fs,
	}
}

func (cmd *CreateUserProvidedService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("create-user-provided-service"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd *CreateUserProvidedService) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.userProvidedServiceInstanceRepo = deps.RepoLocator.GetUserProvidedServiceInstanceRepository()
	return cmd
}

func (cmd *CreateUserProvidedService) Execute(c flags.FlagContext) {
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
