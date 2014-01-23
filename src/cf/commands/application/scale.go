package application

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/formatters"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type Scale struct {
	ui        terminal.UI
	config    *configuration.Configuration
	restarter ApplicationRestarter
	appReq    requirements.ApplicationRequirement
	appRepo   api.ApplicationRepository
}

func NewScale(ui terminal.UI, config *configuration.Configuration, restarter ApplicationRestarter, appRepo api.ApplicationRepository) (cmd *Scale) {
	cmd = new(Scale)
	cmd.ui = ui
	cmd.config = config
	cmd.restarter = restarter
	cmd.appRepo = appRepo
	return
}

func (cmd *Scale) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {

	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "scale")
		return
	}

	if c.Int("i") == -1 && c.String("m") == "" {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "scale")
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

func (cmd *Scale) Run(c *cli.Context) {
	currentApp := cmd.appReq.GetApplication()
	cmd.ui.Say("Scaling app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(currentApp.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.SpaceFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	params := cf.NewEmptyAppParams()
	shouldRestart := false

	if c.String("m") != "" {
		memory, err := formatters.ToMegabytes(c.String("m"))
		if err != nil {
			cmd.ui.Say("Invalid value for memory")
			cmd.ui.FailWithUsage(c, "scale")
			return
		}
		params.Set("memory", memory)
		shouldRestart = true
	}

	if c.Int("i") != -1 {
		params.Set("instances", c.Int("i"))
	}

	updatedApp, apiResponse := cmd.appRepo.Update(currentApp.Guid, params)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if shouldRestart {
		cmd.restarter.ApplicationRestart(updatedApp)
	}
}
