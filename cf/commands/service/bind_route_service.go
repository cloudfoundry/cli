package service

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/flagcontext"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type BindRouteService struct {
	ui                      terminal.UI
	config                  coreconfig.Reader
	routeRepo               api.RouteRepository
	routeServiceBindingRepo api.RouteServiceBindingRepository
	domainReq               requirements.DomainRequirement
	serviceInstanceReq      requirements.ServiceInstanceRequirement
}

func init() {
	commandregistry.Register(&BindRouteService{})
}

func (cmd *BindRouteService) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["hostname"] = &flags.StringFlag{
		Name:      "hostname",
		ShortName: "n",
		Usage:     T("Hostname used in combination with DOMAIN to specify the route to bind"),
	}
	fs["path"] = &flags.StringFlag{
		Name:  "path",
		Usage: T("Path for the HTTP route"),
	}
	fs["parameters"] = &flags.StringFlag{
		ShortName: "c",
		Usage:     T("Valid JSON object containing service-specific configuration parameters, provided inline or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."),
	}
	fs["f"] = &flags.BackwardsCompatibilityFlag{}

	return commandregistry.CommandMetadata{
		Name:        "bind-route-service",
		ShortName:   "brs",
		Description: T("Bind a service instance to an HTTP route"),
		Usage: []string{
			T(`CF_NAME bind-route-service DOMAIN SERVICE_INSTANCE [--hostname HOSTNAME] [--path PATH] [-c PARAMETERS_AS_JSON]`),
		},
		Examples: []string{
			`CF_NAME bind-route-service example.com myratelimiter --hostname myapp --path foo`,
			`CF_NAME bind-route-service example.com myratelimiter -c file.json`,
			`CF_NAME bind-route-service example.com myratelimiter -c '{"valid":"json"}'`,
			``,
			T(`In Windows PowerShell use double-quoted, escaped JSON: "{\"valid\":\"json\"}"`),
			T(`In Windows Command Line use single-quoted, escaped JSON: '{\"valid\":\"json\"}'`),
		},
		Flags: fs,
	}
}

func (cmd *BindRouteService) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires DOMAIN and SERVICE_INSTANCE as arguments\n\n") + commandregistry.Commands.CommandUsage("bind-route-service"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	domainName := fc.Args()[0]
	cmd.domainReq = requirementsFactory.NewDomainRequirement(domainName)

	serviceName := fc.Args()[1]
	cmd.serviceInstanceReq = requirementsFactory.NewServiceInstanceRequirement(serviceName)

	minAPIVersionRequirement := requirementsFactory.NewMinAPIVersionRequirement(
		"bind-route-service",
		cf.MultipleAppPortsMinimumAPIVersion,
	)

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.domainReq,
		cmd.serviceInstanceReq,
		minAPIVersionRequirement,
	}
	return reqs, nil
}

func (cmd *BindRouteService) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	cmd.routeServiceBindingRepo = deps.RepoLocator.GetRouteServiceBindingRepository()
	return cmd
}

func (cmd *BindRouteService) Execute(c flags.FlagContext) error {
	var port int

	host := c.String("hostname")
	domain := cmd.domainReq.GetDomain()
	path := c.String("path")
	if !strings.HasPrefix(path, "/") && len(path) > 0 {
		path = fmt.Sprintf("/%s", path)
	}

	var parameters string

	if c.IsSet("parameters") {
		jsonBytes, err := flagcontext.GetContentsFromFlagValue(c.String("parameters"))
		if err != nil {
			return err
		}
		parameters = string(jsonBytes)
	}

	route, err := cmd.routeRepo.Find(host, domain, path, port)
	if err != nil {
		return err
	}

	serviceInstance := cmd.serviceInstanceReq.GetServiceInstance()

	cmd.ui.Say(T("Binding route {{.URL}} to service instance {{.ServiceInstanceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"ServiceInstanceName": terminal.EntityNameColor(serviceInstance.Name),
			"URL":         terminal.EntityNameColor(route.URL()),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	err = cmd.routeServiceBindingRepo.Bind(serviceInstance.GUID, route.GUID, serviceInstance.IsUserProvided(), parameters)
	if err != nil {
		if httpErr, ok := err.(errors.HTTPError); ok && httpErr.ErrorCode() == errors.ServiceInstanceAlreadyBoundToSameRoute {
			cmd.ui.Warn(T("Route {{.URL}} is already bound to service instance {{.ServiceInstanceName}}.",
				map[string]interface{}{
					"URL": route.URL(),
					"ServiceInstanceName": serviceInstance.Name,
				}))
		} else {
			return err
		}
	}

	cmd.ui.Ok()
	return nil
}
