package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
)

type Delete struct {
	ui      term.UI
	config  *configuration.Configuration
	appRepo api.ApplicationRepository
	appReq  requirements.ApplicationRequirement
}

func NewDelete(ui term.UI, config *configuration.Configuration, appRepo api.ApplicationRepository) (d *Delete) {
	d = new(Delete)
	d.ui = ui
	d.config = config
	d.appRepo = appRepo
	return
}

func (cmd *Delete) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []Requirement, err error) {
	if len(c.Args()) == 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete")
		return
	}
	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []Requirement{&cmd.appReq}
	return
}

func (d *Delete) Run(c *cli.Context) {
	app := d.appReq.Application
	force := c.Bool("f")

	if !force {
		response := strings.ToLower(d.ui.Ask("Really delete %s?>", app.Name))
		if response != "y" && response != "yes" {
			return
		}
	}

	d.ui.Say("Deleting %s", app.Name)
	err := d.appRepo.Delete(d.config, app)
	if err != nil {
		d.ui.Failed("Error deleting app.", err)
		return
	}

	d.ui.Ok()
	return
}
