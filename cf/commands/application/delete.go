package application

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteApp struct {
	ui        terminal.UI
	config    configuration.Reader
	appRepo   api.ApplicationRepository
	routeRepo api.RouteRepository
	appReq    requirements.ApplicationRequirement
}

func NewDeleteApp(ui terminal.UI, config configuration.Reader, appRepo api.ApplicationRepository, routeRepo api.RouteRepository) (cmd *DeleteApp) {
	cmd = new(DeleteApp)
	cmd.ui = ui
	cmd.config = config
	cmd.appRepo = appRepo
	cmd.routeRepo = routeRepo
	return
}

func (cmd *DeleteApp) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete",
		ShortName:   "d",
		Description: T("Delete an app"),
		Usage:       T("CF_NAME delete APP [-f -r]"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")},
			cli.BoolFlag{Name: "r", Usage: T("Also delete any mapped routes")},
		},
	}
}

func (cmd *DeleteApp) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return
}

func (cmd *DeleteApp) Run(c *cli.Context) {
	appName := c.Args()[0]

	if !c.Bool("f") {
		response := cmd.ui.ConfirmDelete(T("app"), appName)
		if !response {
			return
		}
	}

	cmd.ui.Say(T("Deleting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(appName),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	app, apiErr := cmd.appRepo.Read(appName)

	switch apiErr.(type) {
	case nil: // no error
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("App {{.AppName}} does not exist.", map[string]interface{}{"AppName": appName}))
		return
	default:
		cmd.ui.Failed(apiErr.Error())
	}

	if c.Bool("r") {
		for _, route := range app.Routes {
			apiErr = cmd.routeRepo.Delete(route.Guid)
			if apiErr != nil {
				cmd.ui.Failed(apiErr.Error())
			}
		}
	}

	apiErr = cmd.appRepo.Delete(app.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	cmd.ui.Ok()
}
