package route

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/command"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type UpdateRoute struct {
	ui         terminal.UI
	config     coreconfig.Reader
	routeRepo  api.RouteRepository
	domainRepo api.DomainRepository
	domainReq  requirements.DomainRequirement
}

func init() {
	commandregistry.Register(&UpdateRoute{})
}

func (cmd *UpdateRoute) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["hostname"] = &flags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname for the HTTP route (required for shared domains)")}
	fs["path"] = &flags.StringFlag{Name: "path", Usage: T("Path for the HTTP route")}
	fs["option"] = &flags.StringFlag{Name: "option", ShortName: "o", Usage: T("Set the value of a per-route option")}
	fs["remove-option"] = &flags.StringFlag{Name: "remove-option", ShortName: "r", Usage: T("Remove an option with the given name")}

	return commandregistry.CommandMetadata{
		Name:        "update-route",
		Description: T("Update an existing HTTP route"),
		Usage: []string{
			fmt.Sprintf("%s:\n", T("Update an existing HTTP route")),
			"      CF_NAME update-route ",
			fmt.Sprintf("%s ", T("DOMAIN")),
			fmt.Sprintf("[--hostname %s] ", T("HOSTNAME")),
			fmt.Sprintf("[--path %s] ", T("PATH")),
			fmt.Sprintf("[--option %s=%s] ", T("OPTION"), T("VALUE")),
			fmt.Sprintf("[--remove-option %s]\n\n", T("OPTION")),
		},
		Examples: []string{
			"CF_NAME update-route example.com -o loadbalancing=round-robin",
			"CF_NAME update-route example.com -o loadbalancing=least-connection",
			"CF_NAME update-route example.com -r loadbalancing",
			"CF_NAME update-route example.com --hostname myhost --path foo -o loadbalancing=round-robin",
		},
		Flags: fs,
	}
}

func (cmd *UpdateRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires DOMAIN as an argument\n\n") + commandregistry.Commands.CommandUsage("update-route"))
		return nil, fmt.Errorf("incorrect usage: %d arguments of %d required", len(fc.Args()), 1)
	}

	domainName := fc.Args()[0]
	cmd.domainReq = requirementsFactory.NewDomainRequirement(domainName)

	var reqs []requirements.Requirement

	reqs = append(reqs, []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.domainReq,
	}...)

	return reqs, nil
}

func (cmd *UpdateRoute) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()

	return cmd
}

func (cmd *UpdateRoute) Execute(c flags.FlagContext) error {
	domain := c.Args()[0]
	rawHostNameFromFlag := c.String("n")
	host := strings.ToLower(rawHostNameFromFlag)
	path := c.String("path")
	domainFields := cmd.domainReq.GetDomain()
	option := c.String("o")
	removeOption := c.String("r")
	port := 0

	if option != "" {
		err := command.MinimumCCAPIVersionCheck(cmd.config.APIVersion(), ccversion.MinVersionPerRouteOpts)
		if err != nil {
			cmd.ui.Say(T("Your CC API version ({{.APIVersion}}) does not support per route options."+
				"Upgrade to a newer version of the API (minimum version {{.MinSupportedVersion}}). ", map[string]interface{}{
				"APIVersion":          cmd.config.APIVersion(),
				"MinSupportedVersion": ccversion.MinVersionPerRouteOpts,
			}))
		}
	}

	url := (&models.RoutePresenter{
		Host:   host,
		Domain: domain,
		Path:   path,
		Port:   port,
	}).URL()

	route, err := cmd.routeRepo.Find(host, domainFields, path, 0)
	if err != nil {
		cmd.ui.Failed(T("Route with domain '{{.URL}}' does not exist.",
			map[string]interface{}{"URL": url}))
		return err
	}

	err = command.MinimumCCAPIVersionCheck(cmd.config.APIVersion(), ccversion.MinVersionPerRouteOpts)
	if err != nil {
		cmd.ui.Say(T("Your CC API version ({{.APIVersion}}) does not support per route options."+
			"Upgrade to a newer version of the API (minimum version {{.MinSupportedVersion}}). ", map[string]interface{}{
			"APIVersion":          cmd.config.APIVersion(),
			"MinSupportedVersion": ccversion.MinVersionPerRouteOpts,
		}))
		return err
	}

	if c.IsSet("o") {
		key, value, found := strings.Cut(option, "=")
		if found {
			cmd.ui.Say(T("Setting route option {{.Option}} for {{.URL}} to {{.Value}}...", map[string]interface{}{
				"Option": terminal.EntityNameColor(key),
				"Value":  terminal.EntityNameColor(value),
				"URL":    terminal.EntityNameColor(url)}))
			if route.Options == nil {
				route.Options = make(map[string]string)
			}
			route.Options[key] = value
		} else {
			cmd.ui.Say(T("Wrong route option format {{.Option}} for {{.URL}}", map[string]interface{}{
				"Option": terminal.FailureColor(option),
				"URL":    terminal.EntityNameColor(url)}))
		}
	}

	if c.IsSet("r") {
		if _, ok := route.Options[removeOption]; ok {
			cmd.ui.Say(T("Removing route option {{.Option}} for {{.URL}}...", map[string]interface{}{
				"Option": terminal.EntityNameColor(removeOption),
				"URL":    terminal.EntityNameColor(url)}))

			delete(route.Options, removeOption)
		} else {
			cmd.ui.Say(T("No route option {{.Option}} for {{.URL}}", map[string]interface{}{
				"Option": terminal.EntityNameColor(removeOption),
				"URL":    terminal.EntityNameColor(url)}))
		}
	}

	cmd.ui.Ok()
	return nil
}