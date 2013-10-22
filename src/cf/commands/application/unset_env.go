package application

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type UnsetEnv struct {
	ui      terminal.UI
	config  *configuration.Configuration
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewUnsetEnv(ui terminal.UI, config *configuration.Configuration, appRepo api.ApplicationRepository) (cmd *UnsetEnv) {
	cmd = new(UnsetEnv)
	cmd.ui = ui
	cmd.config = config
	cmd.appRepo = appRepo
	return
}

func (cmd *UnsetEnv) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) < 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "unset-env")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *UnsetEnv) Run(c *cli.Context) {
	varName := c.Args()[1]
	app := cmd.appReq.GetApplication()

	cmd.ui.Say("Removing env variable %s from app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(varName),
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	envVars := app.EnvironmentVars

	if !envVarFound(varName, envVars) {
		cmd.ui.Ok()
		cmd.ui.Warn("Env variable %s was not set.", varName)
		return
	}

	delete(envVars, varName)

	apiResponse := cmd.appRepo.SetEnv(app, envVars)

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("TIP: Use '%s push' to ensure your env variable changes take effect", cf.Name)
}
