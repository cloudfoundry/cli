package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf/flagcontext"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type UpdateUserProvidedService struct {
	ui                              terminal.UI
	config                          coreconfig.Reader
	userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository
	serviceInstanceReq              requirements.ServiceInstanceRequirement
}

func init() {
	commandregistry.Register(&UpdateUserProvidedService{})
}

func (cmd *UpdateUserProvidedService) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["p"] = &flags.StringFlag{ShortName: "p", Usage: T("Credentials, provided inline or in a file, to be exposed in the VCAP_SERVICES environment variable for bound applications")}
	fs["l"] = &flags.StringFlag{ShortName: "l", Usage: T("URL to which logs for bound applications will be streamed")}
	fs["r"] = &flags.StringFlag{ShortName: "r", Usage: T("URL to which requests for bound routes will be forwarded. Scheme for this URL must be https")}

	return commandregistry.CommandMetadata{
		Name:        "update-user-provided-service",
		ShortName:   "uups",
		Description: T("Update user-provided service instance"),
		Usage: []string{
			T(`CF_NAME update-user-provided-service SERVICE_INSTANCE [-p CREDENTIALS] [-l SYSLOG_DRAIN_URL] [-r ROUTE_SERVICE_URL]

   Pass comma separated credential parameter names to enable interactive mode:
   CF_NAME update-user-provided-service SERVICE_INSTANCE -p "comma, separated, parameter, names"

   Pass credential parameters as JSON to create a service non-interactively:
   CF_NAME update-user-provided-service SERVICE_INSTANCE -p '{"key1":"value1","key2":"value2"}'

   Specify a path to a file containing JSON:
   CF_NAME update-user-provided-service SERVICE_INSTANCE -p PATH_TO_FILE`),
		},
		Examples: []string{
			`CF_NAME update-user-provided-service my-db-mine -p '{"username":"admin", "password":"pa55woRD"}'`,
			"CF_NAME update-user-provided-service my-db-mine -p /path/to/credentials.json",
			"CF_NAME update-user-provided-service my-drain-service -l syslog://example.com",
			"CF_NAME update-user-provided-service my-route-service -r https://example.com",
		},
		Flags: fs,
	}
}

func (cmd *UpdateUserProvidedService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("update-user-provided-service"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.serviceInstanceReq,
	}

	if fc.IsSet("r") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '-r'", cf.MultipleAppPortsMinimumAPIVersion))
	}

	return reqs, nil
}

func (cmd *UpdateUserProvidedService) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.userProvidedServiceInstanceRepo = deps.RepoLocator.GetUserProvidedServiceInstanceRepository()
	return cmd
}

func (cmd *UpdateUserProvidedService) Execute(c flags.FlagContext) error {
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()
	if !serviceInstance.IsUserProvided() {
		return errors.New(T("Service Instance is not user provided"))
	}

	drainURL := c.String("l")
	credentials := strings.Trim(c.String("p"), `'"`)
	routeServiceURL := c.String("r")

	credentialsMap := make(map[string]interface{})

	if c.IsSet("p") {
		jsonBytes, err := flagcontext.GetContentsFromFlagValue(credentials)
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

	cmd.ui.Say(T("Updating user provided service {{.ServiceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceName": terminal.EntityNameColor(serviceInstance.Name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	serviceInstance.Params = credentialsMap
	serviceInstance.SysLogDrainURL = drainURL
	serviceInstance.RouteServiceURL = routeServiceURL

	err := cmd.userProvidedServiceInstanceRepo.Update(serviceInstance.ServiceInstanceFields)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("TIP: Use '{{.CFRestageCommand}}' for any bound apps to ensure your env variable changes take effect",
		map[string]interface{}{
			"CFRestageCommand": terminal.CommandColor(cf.Name + " restage"),
		}))

	if routeServiceURL == "" && credentials == "" && drainURL == "" {
		cmd.ui.Warn(T("No flags specified. No changes were made."))
	}
	return nil
}
