package service

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

//go:generate counterfeiter . RouteServiceUnbinder

type RouteServiceUnbinder interface {
	UnbindRoute(route models.Route, serviceInstance models.ServiceInstance) error
}

type UnbindRouteService struct {
	ui                      terminal.UI
	config                  coreconfig.Reader
	routeRepo               api.RouteRepository
	routeServiceBindingRepo api.RouteServiceBindingRepository
	domainReq               requirements.DomainRequirement
	serviceInstanceReq      requirements.ServiceInstanceRequirement
}

func init() {
	commandregistry.Register(&UnbindRouteService{})
}

func (cmd *UnbindRouteService) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["hostname"] = &flags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname used in combination with DOMAIN to specify the route to unbind")}
	fs["path"] = &flags.StringFlag{Name: "path", Usage: T("Path for HTTP route")}
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force unbinding without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "unbind-route-service",
		ShortName:   "urs",
		Description: T("Unbind a service instance from an HTTP route"),
		Usage: []string{
			T("CF_NAME unbind-route-service DOMAIN SERVICE_INSTANCE [--hostname HOSTNAME] [--path PATH] [-f]"),
		},
		Examples: []string{
			"CF_NAME unbind-route-service example.com myratelimiter --hostname myapp --path foo",
		},
		Flags: fs,
	}
}

func (cmd *UnbindRouteService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires DOMAIN and SERVICE_INSTANCE as arguments\n\n") + commandregistry.Commands.CommandUsage("unbind-route-service"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	serviceName := fc.Args()[1]
	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(serviceName)

	domainName := fc.Args()[0]
	cmd.domainReq = requirementsFactory.NewDomainRequirement(domainName)

	minAPIVersionRequirement := requirementsFactory.NewMinAPIVersionRequirement(
		"unbind-route-service",
		cf.MultipleAppPortsMinimumAPIVersion,
	)

	reqs := []requirements.Requirement{
		minAPIVersionRequirement,
		requirementsFactory.NewLoginRequirement(),
		cmd.domainReq,
		cmd.serviceInstanceReq,
	}
	return reqs, nil
}

func (cmd *UnbindRouteService) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	cmd.routeServiceBindingRepo = deps.RepoLocator.GetRouteServiceBindingRepository()
	return cmd
}

func (cmd *UnbindRouteService) Execute(c flags.FlagContext) error {
	var port int

	host := c.String("hostname")
	domain := cmd.domainReq.GetDomain()
	path := c.String("path")
	if !strings.HasPrefix(path, "/") && len(path) > 0 {
		path = fmt.Sprintf("/%s", path)
	}

	route, err := cmd.routeRepo.Find(host, domain, path, port)
	if err != nil {
		return err
	}

	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()
	confirmed := c.Bool("f")
	if !confirmed {
		confirmed = cmd.ui.Confirm(T("Unbinding may leave apps mapped to route {{.URL}} vulnerable; e.g. if service instance {{.ServiceInstanceName}} provides authentication. Do you want to proceed?",
			map[string]interface{}{
				"URL": route.URL(),
				"ServiceInstanceName": serviceInstance.Name,
			}))

		if !confirmed {
			cmd.ui.Warn(T("Unbind cancelled"))
			return nil
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
		httpError, ok := err.(errors.HTTPError)
		if ok && httpError.ErrorCode() == errors.InvalidRelation {
			cmd.ui.Warn(T("Route {{.Route}} was not bound to service instance {{.ServiceInstance}}.", map[string]interface{}{"Route": route.URL(), "ServiceInstance": serviceInstance.Name}))
		} else {
			return err
		}
	}

	cmd.ui.Ok()
	return nil
}

func (cmd *UnbindRouteService) UnbindRoute(route models.Route, serviceInstance models.ServiceInstance) error {
	return cmd.routeServiceBindingRepo.Unbind(serviceInstance.GUID, route.GUID, serviceInstance.IsUserProvided())
}
