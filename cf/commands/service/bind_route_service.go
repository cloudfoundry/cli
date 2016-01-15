package service

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type RouteServiceBinder interface {
	BindRoute(route models.Route, serviceInstance models.ServiceInstance) (apiErr error)
}

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
	baseUsage := T("CF_NAME bind-route-service DOMAIN SERVICE_INSTANCE [-n HOST] [-f]")
	exampleUsage := T(`EXAMPLE:
   CF_NAME bind-route-service example.com myratelimiter -n myapp`)

	fs := make(map[string]flags.FlagSet)
	fs["n"] = &cliFlags.StringFlag{ShortName: "n", Usage: T("Hostname used in combination with DOMAIN to specify the route to bind")}
	fs["f"] = &cliFlags.BoolFlag{ShortName: "f", Usage: T("Force binding without confirmation")}

	return command_registry.CommandMetadata{
		Name:        "bind-route-service",
		ShortName:   "brs",
		Description: T("Bind a service instance to a route"),
		Usage:       strings.Join([]string{baseUsage, exampleUsage}, "\n\n"),
		Flags:       fs,
	}
}

func (cmd *BindRouteService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires DOMAIN and SERVICE_INSTANCE as arguments\n\n") + command_registry.Commands.CommandUsage("bind-route-service"))
	}

	serviceName := fc.Args()[1]

	cmd.domainReq = requirementsFactory.NewDomainRequirement(fc.Args()[0])
	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(serviceName)

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement(), cmd.domainReq, cmd.serviceInstanceReq}

	return
}

func (cmd *BindRouteService) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	cmd.routeServiceBindingRepo = deps.RepoLocator.GetRouteServiceBindingRepository()
	return cmd
}

func (cmd *BindRouteService) Execute(c flags.FlagContext) {
	domain := cmd.domainReq.GetDomain()
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()

	host := c.String("n")

	route, err := cmd.routeRepo.Find(host, domain, "")

	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	if !serviceInstance.IsUserProvided() {
		confirmed := c.Bool("f")
		requiresRouteForwarding := requiresRouteForwarding(serviceInstance)

		if requiresRouteForwarding && !confirmed {
			confirmed = cmd.ui.Confirm(T("Binding may cause requests for route {{.URL}} to be altered by service instance {{.ServiceInstanceName}}. Do you want to proceed? (y/n)",
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

	err = cmd.BindRoute(route, serviceInstance)
	if err != nil {
		if httperr, ok := err.(errors.HttpError); ok && httperr.ErrorCode() == errors.ROUTE_ALREADY_BOUND_TO_SAME_SERVICE {
			cmd.ui.Ok()
			cmd.ui.Warn(T("Route {{.URL}} is already bound to service instance {{.ServiceInstanceName}}.",
				map[string]interface{}{
					"URL": route.URL(),
					"ServiceInstanceName": serviceInstance.Name,
				}))
			return
		} else {
			cmd.ui.Failed(err.Error())
		}
	}

	cmd.ui.Ok()
}

func (cmd *BindRouteService) BindRoute(route models.Route, serviceInstance models.ServiceInstance) error {
	return cmd.routeServiceBindingRepo.Bind(serviceInstance.Guid, route.Guid, serviceInstance.IsUserProvided())
}

func requiresRouteForwarding(serviceInstance models.ServiceInstance) bool {
	for _, require := range serviceInstance.ServiceOffering.Requires {
		if require == "route_forwarding" {
			return true
		}
	}
	return false
}
