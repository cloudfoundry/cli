package application

import (
	"cf/api"
	"cf/configuration"
	"cf/errors"
	"cf/requirements"
	"cf/terminal"
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

func (cmd *DeleteApp) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete")
		return
	}

	reqs = []requirements.Requirement{reqFactory.NewLoginRequirement()}
	return
}

func (cmd *DeleteApp) Run(c *cli.Context) {
	appName := c.Args()[0]

	if !c.Bool("f") {
		response := cmd.ui.ConfirmDelete("app", appName)
		if !response {
			return
		}
	}

	cmd.ui.Say("Deleting app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(appName),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	app, apiErr := cmd.appRepo.Read(appName)

	switch apiErr.(type) {
	case nil: // no error
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn("App %s does not exist.", appName)
		return
	default:
		cmd.ui.Failed(apiErr.Error())
		return
	}

	if c.Bool("r") {
		for _, route := range app.Routes {
			apiErr = cmd.routeRepo.Delete(route.Guid)
			if apiErr != nil {
				cmd.ui.Failed(apiErr.Error())
				return
			}
		}
	}

	apiErr = cmd.appRepo.Delete(app.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	return
}
