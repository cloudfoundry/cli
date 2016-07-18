package route

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type DeleteRoute struct {
	ui        terminal.UI
	config    coreconfig.Reader
	routeRepo api.RouteRepository
	domainReq requirements.DomainRequirement
}

func init() {
	commandregistry.Register(&DeleteRoute{})
}

func (cmd *DeleteRoute) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}
	fs["hostname"] = &flags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname used to identify the HTTP route")}
	fs["path"] = &flags.StringFlag{Name: "path", Usage: T("Path used to identify the HTTP route")}
	fs["port"] = &flags.IntFlag{Name: "port", Usage: T("Port used to identify the TCP route")}

	return commandregistry.CommandMetadata{
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
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("delete-route"))
	}

	if fc.IsSet("port") && (fc.IsSet("hostname") || fc.IsSet("path")) {
		cmd.ui.Failed(T("Cannot specify port together with hostname and/or path."))
	}

	cmd.domainReq = requirementsFactory.NewDomainRequirement(fc.Args()[0])

	var reqs []requirements.Requirement

	if fc.String("path") != "" {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--path'", cf.RoutePathMinimumAPIVersion))
	}

	if fc.IsSet("port") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--port'", cf.TCPRoutingMinimumAPIVersion))
	}

	reqs = append(reqs, []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.domainReq,
	}...)

	return reqs
}

func (cmd *DeleteRoute) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	return cmd
}

func (cmd *DeleteRoute) Execute(c flags.FlagContext) error {
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
			return nil
		}
	}

	cmd.ui.Say(T("Deleting route {{.URL}}...", map[string]interface{}{"URL": terminal.EntityNameColor(url)}))

	domain := cmd.domainReq.GetDomain()
	route, err := cmd.routeRepo.Find(host, domain, path, port)
	if err != nil {
		if _, ok := err.(*errors.ModelNotFoundError); ok {
			cmd.ui.Warn(T("Unable to delete, route '{{.URL}}' does not exist.",
				map[string]interface{}{"URL": url}))
			return nil
		}
		return err
	}

	err = cmd.routeRepo.Delete(route.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
