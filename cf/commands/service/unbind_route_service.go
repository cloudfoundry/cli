package service

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/simonleung8/flags"
	"github.com/simonleung8/flags/flag"
)

type RouteServiceUnbinder interface {
	UnbindRoute(route models.Route, serviceInstance models.ServiceInstance) (apiErr error)
}

type UnbindRouteService struct {
	ui                      terminal.UI
	config                  core_config.Reader
	routeRepo               api.RouteRepository
	routeServiceBindingRepo api.RouteServiceBindingRepository
	domainReq               requirements.DomainRequirement
	serviceInstanceReq      requirements.ServiceInstanceRequirement
}

func init() {
	command_registry.Register(&UnbindRouteService{})
}

func (cmd *UnbindRouteService) MetaData() command_registry.CommandMetadata {
	baseUsage := T("CF_NAME unbind-route-service DOMAIN SERVICE_INSTANCE [-n HOST] [-f]")
	exampleUsage := T(`EXAMPLE:
		CF_NAME unbind-route-service 10.244.0.34.xip.io myratelimiter -n spring-music`)

	fs := make(map[string]flags.FlagSet)
	fs["n"] = &cliFlags.StringFlag{Name: "n", Usage: T("Hostname used in combination with DOMAIN to specify the route to unbind")}
	fs["f"] = &cliFlags.BoolFlag{Name: "f", Usage: T("Force unbinding without confirmation")}

	return command_registry.CommandMetadata{
		Name:        "unbind-route-service",
		ShortName:   "ubrs",
		Description: T("Unbind a service instance from a route"),
		Usage:       strings.Join([]string{baseUsage, exampleUsage}, "\n\n"),
		Flags:       fs,
	}
}

func (cmd *UnbindRouteService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires DOMAIN and SERVICE_INSTANCE as arguments\n\n") + command_registry.Commands.CommandUsage("unbind-route-service"))
	}

	serviceName := fc.Args()[1]

	cmd.domainReq = requirementsFactory.NewDomainRequirement(fc.Args()[0])
	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(serviceName)

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement(), cmd.domainReq, cmd.serviceInstanceReq}

	return
}

func (cmd *UnbindRouteService) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	cmd.routeServiceBindingRepo = deps.RepoLocator.GetRouteServiceBindingRepository()
	return cmd
}

func (cmd *UnbindRouteService) Execute(c flags.FlagContext) {
	domain := cmd.domainReq.GetDomain()
	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()

	host := c.String("n")

	route, err := cmd.routeRepo.FindByHostAndDomain(host, domain)

	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	confirmed := c.Bool("f")

	if !confirmed {
		confirmed = cmd.ui.Confirm(T("Unbinding may leave apps mapped to route {{.URL}} vulnerable; e.g. if service instance {{.ServiceInstanceName}} provides authentication. Do you want to proceed? (y/n)",
			map[string]interface{}{
				"URL": route.URL(),
				"ServiceInstanceName": serviceInstance.Name,
			}))

		if !confirmed {
			return
		}
	}

	cmd.ui.Say(T("Unbinding route {{.URL}} from service instance {{.ServiceInstanceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name),
			"URL":         terminal.EntityNameColor(route.URL()),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err = cmd.UnbindRoute(route, serviceInstance)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}

func (cmd *UnbindRouteService) UnbindRoute(route models.Route, serviceInstance models.ServiceInstance) error {
	return cmd.routeServiceBindingRepo.Unbind(serviceInstance.Guid, route.Guid, serviceInstance.IsUserProvided())
}
