package service

import (
	"encoding/json"
	"strings"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/util"

	"fmt"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type CreateUserProvidedService struct {
	ui                              terminal.UI
	config                          coreconfig.Reader
	userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository
}

func init() {
	commandregistry.Register(&CreateUserProvidedService{})
}

func (cmd *CreateUserProvidedService) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["p"] = &flags.StringFlag{ShortName: "p", Usage: T("Credentials, provided inline or in a file, to be exposed in the VCAP_SERVICES environment variable for bound applications")}
	fs["l"] = &flags.StringFlag{ShortName: "l", Usage: T("URL to which logs for bound applications will be streamed")}
	fs["r"] = &flags.StringFlag{ShortName: "r", Usage: T("URL to which requests for bound routes will be forwarded. Scheme for this URL must be https")}

	return commandregistry.CommandMetadata{
		Name:        "create-user-provided-service",
		ShortName:   "cups",
		Description: T("Make a user-provided service instance available to CF apps"),
		Usage: []string{
			T(`CF_NAME create-user-provided-service SERVICE_INSTANCE [-p CREDENTIALS] [-l SYSLOG_DRAIN_URL] [-r ROUTE_SERVICE_URL]

   Pass comma separated credential parameter names to enable interactive mode:
   CF_NAME create-user-provided-service SERVICE_INSTANCE -p "comma, separated, parameter, names"

   Pass credential parameters as JSON to create a service non-interactively:
   CF_NAME create-user-provided-service SERVICE_INSTANCE -p '{"key1":"value1","key2":"value2"}'

   Specify a path to a file containing JSON:
   CF_NAME create-user-provided-service SERVICE_INSTANCE -p PATH_TO_FILE`),
		},
		Examples: []string{
			`CF_NAME create-user-provided-service my-db-mine -p "username, password"`,
			`CF_NAME create-user-provided-service my-db-mine -p /path/to/credentials.json`,
			`CF_NAME create-user-provided-service my-drain-service -l syslog://example.com`,
			`CF_NAME create-user-provided-service my-route-service -r https://example.com`,
			``,
			fmt.Sprintf("%s:", T(`Linux/Mac`)),
			`   CF_NAME create-user-provided-service my-db-mine -p '{"username":"admin","password":"pa55woRD"}'`,
			``,
			fmt.Sprintf("%s:", T(`Windows Command Line`)),
			`   CF_NAME create-user-provided-service my-db-mine -p "{\"username\":\"admin\",\"password\":\"pa55woRD\"}"`,
			``,
			fmt.Sprintf("%s:", T(`Windows PowerShell`)),
			`   CF_NAME create-user-provided-service my-db-mine -p '{\"username\":\"admin\",\"password\":\"pa55woRD\"}'`,
		},
		Flags: fs,
	}
}

func (cmd *CreateUserProvidedService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("create-user-provided-service"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}

	if fc.IsSet("r") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '-r'", cf.MultipleAppPortsMinimumAPIVersion))
	}

	return reqs
}

func (cmd *CreateUserProvidedService) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.userProvidedServiceInstanceRepo = deps.RepoLocator.GetUserProvidedServiceInstanceRepository()
	return cmd
}

func (cmd *CreateUserProvidedService) Execute(c flags.FlagContext) error {
	name := c.Args()[0]
	drainURL := c.String("l")
	routeServiceURL := c.String("r")
	credentials := strings.Trim(c.String("p"), `"'`)
	credentialsMap := make(map[string]interface{})

	if c.IsSet("p") {
		jsonBytes, err := util.GetContentsFromFlagValue(credentials)
		if err != nil {
			return err
		}

		err = json.Unmarshal(jsonBytes, &credentialsMap)
		if err != nil {
			for _, param := range strings.Split(credentials, ",") {
				param = strings.Trim(param, " ")
				credentialsMap[param] = cmd.ui.Ask(param)
			}
		}
	}

	cmd.ui.Say(T("Creating user provided service {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceName": terminal.EntityNameColor(name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err := cmd.userProvidedServiceInstanceRepo.Create(name, drainURL, routeServiceURL, credentialsMap)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
