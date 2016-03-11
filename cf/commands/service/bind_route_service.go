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
	fs["force"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force binding without confirmation")}
	fs["hostname"] = &flags.StringFlag{
		Name:      "hostname",
		ShortName: "n",
		Usage:     T("Hostname used in combination with DOMAIN to specify the route to bind"),
	}
	fs["parameters"] = &flags.StringFlag{
		ShortName: "c",
		Usage:     T("Valid JSON object containing service-specific configuration parameters, provided inline or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."),
	}

	return command_registry.CommandMetadata{
		Name:        "bind-route-service",
		ShortName:   "brs",
		Description: T("Bind a service instance to an HTTP route"),
		Usage: []string{
			T(`CF_NAME bind-route-service DOMAIN SERVICE_INSTANCE [-f] [--hostname HOSTNAME] [-c PARAMETERS_AS_JSON]`),
		},
		Examples: []string{
			`CF_NAME bind-route-service example.com myratelimiter --hostname myapp`,
			`CF_NAME bind-route-service example.com myratelimiter -c file.json`,
			`CF_NAME bind-route-service example.com myratelimiter -c '{"valid":"json"}'`,
			``,
			T(`In Windows PowerShell use double-quoted, escaped JSON: "{\"valid\":\"json\"}"`),
			T(`In Windows Command Line use single-quoted, escaped JSON: '{\"valid\":\"json\"}'`),
		},
		Flags: fs,
	}
}

func (cmd *BindRouteService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
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
		"bind-route-service",
		minAPIVersion,
	)

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.domainReq,
		cmd.serviceInstanceReq,
		minAPIVersionRequirement,
	}
	return reqs
}

func (cmd *BindRouteService) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	cmd.routeServiceBindingRepo = deps.RepoLocator.GetRouteServiceBindingRepository()
	return cmd
}

func (cmd *BindRouteService) Execute(c flags.FlagContext) {
	var path string
	var port int

	host := c.String("hostname")
	domain := cmd.domainReq.GetDomain()

	var parameters string

	if c.IsSet("parameters") {
		jsonBytes, err := util.GetContentsFromFlagValue(c.String("parameters"))
		if err != nil {
			cmd.ui.Failed(err.Error())
		}
		parameters = string(jsonBytes)
	}

	route, err := cmd.routeRepo.Find(host, domain, path, port)
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
		if httpErr, ok := err.(errors.HttpError); ok && httpErr.ErrorCode() == errors.ServiceInstanceAlreadyBoundToSameRoute {
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
