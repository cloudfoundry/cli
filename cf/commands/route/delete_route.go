package route

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type DeleteRoute struct {
	ui        terminal.UI
	config    core_config.Reader
	routeRepo api.RouteRepository
	domainReq requirements.DomainRequirement
}

func init() {
	command_registry.Register(&DeleteRoute{})
}

func (cmd *DeleteRoute) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &cliFlags.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")}
	fs["hostname"] = &cliFlags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname")}

	return command_registry.CommandMetadata{
		Name:        "delete-route",
		Description: T("Delete a route"),
		Usage:       T("CF_NAME delete-route DOMAIN [--hostname HOSTNAME] [-f]"),
		Flags:       fs,
	}
}

func (cmd *DeleteRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("delete-route"))
	}

	cmd.domainReq = requirementsFactory.NewDomainRequirement(fc.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.domainReq,
	}
	return
}

func (cmd *DeleteRoute) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	return cmd
}

func (cmd *DeleteRoute) Execute(c flags.FlagContext) {
	host := c.String("n")
	path := "" // path is not yet supported
	domainName := c.Args()[0]

	url := domainName
	if host != "" {
		url = host + "." + domainName
	}
	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete("route", url) {
			return
		}
	}

	cmd.ui.Say(T("Deleting route {{.URL}}...", map[string]interface{}{"URL": terminal.EntityNameColor(url)}))

	domain := cmd.domainReq.GetDomain()
	route, apiErr := cmd.routeRepo.Find(host, domain, path)

	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Warn(T("Unable to delete, route '{{.URL}}' does not exist.",
			map[string]interface{}{"URL": url}))
		return
	default:
		cmd.ui.Failed(apiErr.Error())
		return
	}

	apiErr = cmd.routeRepo.Delete(route.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
