package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type Push struct {
	ui         term.UI
	config     *configuration.Configuration
	appRepo    api.ApplicationRepository
	domainRepo api.DomainRepository
	routeRepo  api.RouteRepository
}

func NewPush(ui term.UI, config *configuration.Configuration, aR api.ApplicationRepository, dR api.DomainRepository, rR api.RouteRepository) (p Push) {
	p.ui = ui
	p.config = config
	p.appRepo = aR
	p.domainRepo = dR
	p.routeRepo = rR
	return
}

func (p Push) Run(c *cli.Context) {
	appName := c.String("name")
	app, err := p.appRepo.FindByName(p.config, appName)

	if err != nil {
		app, err = p.createApp(p.config, appName, c)

		if err != nil {
			return
		}
	}

	p.ui.Say("Uploading %s...", app.Name)
	err = p.appRepo.Upload(p.config, app)
	if err != nil {
		p.ui.Failed("Error uploading app", err)
		return
	}
	p.ui.Ok()
}

func (p Push) createApp(config *configuration.Configuration, appName string, c *cli.Context) (app cf.Application, err error) {
	newApp := cf.Application{Name: appName}

	p.ui.Say("Creating %s...", appName)
	app, err = p.appRepo.Create(config, newApp)
	if err != nil {
		p.ui.Failed("Error creating application", err)
		return
	}
	p.ui.Ok()

	domain, err := p.domainRepo.FindByName(config, c.String("domain"))

	if err != nil {
		p.ui.Failed("Error loading domain", err)
		return
	}

	newRoute := cf.Route{Host: app.Name}

	p.ui.Say("Creating route %s.%s...", app.Name, domain.Name)
	createdRoute, err := p.routeRepo.Create(config, newRoute, domain)
	if err != nil {
		p.ui.Failed("Error creating route", err)
		return
	}
	p.ui.Ok()

	p.ui.Say("Binding %s.%s to %s...", app.Name, domain.Name, app.Name)
	err = p.routeRepo.Bind(config, createdRoute, app)
	if err != nil {
		p.ui.Failed("Error binding route", err)
		return
	}
	p.ui.Ok()

	return
}
