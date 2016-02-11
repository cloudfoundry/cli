package service

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/util"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type BindRouteService struct {
	ui                      terminal.UI
	config                  core_config.Reader
	routeRepo               api.RouteRepository
	routeServiceBindingRepo api.RouteServiceBindingRepository
	domainReq               requirements.DomainRequirement
	serviceInstanceReq      requirements.ServiceInstanceRequirement
}

func init() {
	command_registry.Register(&BindRouteService{})
}

func (cmd *BindRouteService) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["force"] = &cliFlags.BoolFlag{ShortName: "f", Usage: T("Force binding without confirmation")}
	fs["hostname"] = &cliFlags.StringFlag{
		Name:      "hostname",
		ShortName: "n",
		Usage:     T("Hostname used in combination with DOMAIN to specify the route to bind"),
	}
	fs["parameters"] = &cliFlags.StringFlag{
		ShortName: "c",
		Usage:     T("Valid JSON object containing service-specific configuration parameters, provided inline or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."),
	}

	return command_registry.CommandMetadata{
		Name:        "bind-route-service",
		ShortName:   "brs",
		Description: T("Bind a service instance to a route"),
		Usage: T(`CF_NAME bind-route-service DOMAIN SERVICE_INSTANCE [-f] [--hostname HOSTNAME] [-c PARAMETERS_AS_JSON]

EXAMPLES:
   CF_NAME bind-route-service example.com myratelimiter --hostname myapp
   CF_NAME bind-route-service example.com myratelimiter -c file.json
   CF_NAME bind-route-service example.com myratelimiter -c '{"valid":"json"}'

   In Windows PowerShell use double-quoted, escaped JSON: "{\"valid\":\"json\"}"
   In Windows Command Line use single-quoted, escaped JSON: '{\"valid\":\"json\"}'`),
		Flags: fs,
	}
}

func (cmd *BindRouteService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires DOMAIN and SERVICE_INSTANCE as arguments\n\n") + command_registry.Commands.CommandUsage("bind-route-service"))
	}

	domainName := fc.Args()[0]
	cmd.domainReq = requirementsFactory.NewDomainRequirement(domainName)

	serviceName := fc.Args()[1]
	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(serviceName)

	minAPIVersion, err := semver.Make("2.51.0")
	if err != nil {
		panic(err.Error())
	}

	minAPIVersionRequirement := requirementsFactory.NewMinAPIVersionRequirement(
		T("Option '--parameters/-c'"),
		minAPIVersion,
	)

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.domainReq,
		cmd.serviceInstanceReq,
		minAPIVersionRequirement,
	}, nil
}

func (cmd *BindRouteService) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	cmd.routeServiceBindingRepo = deps.RepoLocator.GetRouteServiceBindingRepository()
	return cmd
}

func (cmd *BindRouteService) Execute(c flags.FlagContext) {
	host := c.String("hostname")
	domain := cmd.domainReq.GetDomain()
	path := "" // path is not currently supported
	var parameters string

	if c.IsSet("parameters") {
		jsonBytes, err := util.GetContentsFromFlagValue(c.String("parameters"))
		if err != nil {
			cmd.ui.Failed(err.Error())
		}
		parameters = string(jsonBytes)
	}

	route, err := cmd.routeRepo.Find(host, domain, path)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()
	if !serviceInstance.IsUserProvided() {
		var requiresRouteForwarding bool
		for _, requirement := range serviceInstance.ServiceOffering.Requires {
			if requirement == "route_forwarding" {
				requiresRouteForwarding = true
				break
			}
		}

		confirmed := c.Bool("force")
		if requiresRouteForwarding && !confirmed {
			confirmed = cmd.ui.Confirm(T("Binding may cause requests for route {{.URL}} to be altered by service instance {{.ServiceInstanceName}}. Do you want to proceed?",
				map[string]interface{}{
					"URL": route.URL(),
					"ServiceInstanceName": serviceInstance.Name,
				}))

			if !confirmed {
				cmd.ui.Warn(T("Bind cancelled"))
				return
			}
		}
	}

	cmd.ui.Say(T("Binding route {{.URL}} to service instance {{.ServiceInstanceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name),
			"URL":         terminal.EntityNameColor(route.URL()),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err = cmd.routeServiceBindingRepo.Bind(serviceInstance.Guid, route.Guid, serviceInstance.IsUserProvided(), parameters)
	if err != nil {
		if httpErr, ok := err.(errors.HttpError); ok && httpErr.ErrorCode() == errors.ROUTE_ALREADY_BOUND_TO_SAME_SERVICE {
			cmd.ui.Warn(T("Route {{.URL}} is already bound to service instance {{.ServiceInstanceName}}.",
				map[string]interface{}{
					"URL": route.URL(),
					"ServiceInstanceName": serviceInstance.Name,
				}))
		} else {
			cmd.ui.Failed(err.Error())
		}
	}

	cmd.ui.Ok()
}
