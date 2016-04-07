package route

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
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
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}
	fs["hostname"] = &flags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname used to identify the route")}
	fs["path"] = &flags.StringFlag{Name: "path", Usage: T("Path used to identify the HTTP route")}
	fs["port"] = &flags.IntFlag{Name: "port", Usage: T("Port used to identify the TCP route")}

	return command_registry.CommandMetadata{
		Name:        "delete-route",
		Description: T("Delete a route"),
		Usage: []string{
			fmt.Sprintf("%s:\n", T("Delete an HTTP route")),
			"      CF_NAME delete-route ",
			fmt.Sprintf("%s ", T("DOMAIN")),
			fmt.Sprintf("[--hostname %s] ", T("HOSTNAME")),
			fmt.Sprintf("[--path %s] [-f]\n\n", T("PATH")),
			fmt.Sprintf("   %s:\n", T("Delete a TCP route")),
			"      CF_NAME delete-route ",
			fmt.Sprintf("%s ", T("DOMAIN")),
			fmt.Sprintf("--port %s [-f]", T("PORT")),
		},
		Examples: []string{
			"CF_NAME delete-route example.com                              # example.com",
			"CF_NAME delete-route example.com --hostname myhost            # myhost.example.com",
			"CF_NAME delete-route example.com --hostname myhost --path foo # myhost.example.com/foo",
			"CF_NAME delete-route example.com --port 50000                 # example.com:50000",
		},
		Flags: fs,
	}
}

func (cmd *DeleteRoute) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("delete-route"))
	}

	if fc.IsSet("port") && (fc.IsSet("hostname") || fc.IsSet("path")) {
		cmd.ui.Failed(T("Cannot specify port together with hostname and/or path."))
	}

	cmd.domainReq = requirementsFactory.NewDomainRequirement(fc.Args()[0])

	requiredVersion, err := semver.Make("2.36.0")
	if err != nil {
		panic(err.Error())
	}

	var reqs []requirements.Requirement

	if fc.String("path") != "" {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--path'", requiredVersion))
	}

	if fc.IsSet("port") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--port'", cf.TcpRoutingMinimumApiVersion))
	}

	reqs = append(reqs, []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.domainReq,
	}...)

	return reqs
}

func (cmd *DeleteRoute) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	return cmd
}

func (cmd *DeleteRoute) Execute(c flags.FlagContext) {
	host := c.String("n")
	path := c.String("path")
	domainName := c.Args()[0]
	port := c.Int("port")

	url := (&models.RoutePresenter{
		Host:   host,
		Domain: domainName,
		Path:   path,
		Port:   port,
	}).URL()

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete("route", url) {
			return
		}
	}

	cmd.ui.Say(T("Deleting route {{.URL}}...", map[string]interface{}{"URL": terminal.EntityNameColor(url)}))

	domain := cmd.domainReq.GetDomain()
	route, err := cmd.routeRepo.Find(host, domain, path, port)
	if err != nil {
		if _, ok := err.(*errors.ModelNotFoundError); ok {
			cmd.ui.Warn(T("Unable to delete, route '{{.URL}}' does not exist.",
				map[string]interface{}{"URL": url}))
			return
		}
		cmd.ui.Failed(err.Error())
		return
	}

	err = cmd.routeRepo.Delete(route.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
}
