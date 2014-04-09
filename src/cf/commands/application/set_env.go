package application

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type SetEnv struct {
	ui      terminal.UI
	config  configuration.Reader
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewSetEnv(ui terminal.UI, config configuration.Reader, appRepo api.ApplicationRepository) (cmd *SetEnv) {
	cmd = new(SetEnv)
	cmd.ui = ui
	cmd.config = config
	cmd.appRepo = appRepo
	return
}

func (cmd *SetEnv) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) < 3 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "set-env")
		return
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *SetEnv) Run(c *cli.Context) {
	varName := c.Args()[1]
	varValue := c.Args()[2]
	app := cmd.appReq.GetApplication()

	cmd.ui.Say("Setting env variable '%s' to '%s' for app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(varName),
		terminal.EntityNameColor(varValue),
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	if len(app.EnvironmentVars) == 0 {
		app.EnvironmentVars = map[string]string{}
	}
	envParams := app.EnvironmentVars
	envParams[varName] = varValue

	_, apiErr := cmd.appRepo.Update(app.Guid, models.AppParams{EnvironmentVars: &envParams})

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("TIP: Use '%s' to ensure your env variable changes take effect", terminal.CommandColor(cf.Name()+" push"))
}
