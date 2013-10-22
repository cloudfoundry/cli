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

type SetEnv struct {
	ui      terminal.UI
	config  *configuration.Configuration
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewSetEnv(ui terminal.UI, config *configuration.Configuration, appRepo api.ApplicationRepository) (cmd *SetEnv) {
	cmd = new(SetEnv)
	cmd.ui = ui
	cmd.config = config
	cmd.appRepo = appRepo
	return
}

func (cmd *SetEnv) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) < 3 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "set-env")
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

func (cmd *SetEnv) Run(c *cli.Context) {
	varName := c.Args()[1]
	varValue := c.Args()[2]
	app := cmd.appReq.GetApplication()

	cmd.ui.Say("Setting env variable %s for app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(varName),
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	var envVars map[string]string

	if app.EnvironmentVars != nil {
		envVars = app.EnvironmentVars
	} else {
		envVars = map[string]string{}
	}

	if envVarFound(varName, envVars) {
		cmd.ui.Ok()
		cmd.ui.Warn("Env var %s was already set.", varName)
		return
	}

	envVars[varName] = varValue

	apiResponse := cmd.appRepo.SetEnv(app, envVars)

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("TIP: Use '%s push' to ensure your env variable changes take effect", cf.Name)
}
