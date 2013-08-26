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
	appRepo    api.ApplicationRepository
	domainRepo api.DomainRepository
	routeRepo  api.RouteRepository
}

func NewPush(ui term.UI, aR api.ApplicationRepository, dR api.DomainRepository, rR api.RouteRepository) (p Push) {
	p.appRepo = aR
	p.domainRepo = dR
	p.routeRepo = rR
	p.ui = ui
	return
}

func (p Push) Run(c *cli.Context) {
	config, err := configuration.Load()

	if err != nil {
		p.ui.Failed("Error loading configuration", err)
		return
	}

	appName := c.String("name")
	newApp := cf.Application{Name: appName}

	p.ui.Say("Creating %s...", appName)
	createdApp, err := p.appRepo.Create(config, newApp)
	if err != nil {
		p.ui.Failed("Error creating application", err)
		return
	}
	p.ui.Ok()

	domains, err := p.domainRepo.FindAll(config)

	if err != nil {
		p.ui.Failed("Error loading domains", err)
		return
	}

	domain := domains[0]
	newRoute := cf.Route{Host: createdApp.Name}

	p.ui.Say("Creating route %s.%s...", createdApp.Name, domain.Name)
	createdRoute, err := p.routeRepo.Create(config, newRoute, domain)
	if err != nil {
		p.ui.Failed("Error creating route", err)
		return
	}
	p.ui.Ok()

	p.ui.Say("Binding %s.%s to %s...", createdApp.Name, domain.Name, createdApp.Name)
	err = p.routeRepo.Bind(config, createdRoute, createdApp)
	if err != nil {
		p.ui.Failed("Error binding route", err)
		return
	}
	p.ui.Ok()

	p.ui.Say("Uploading %s...", createdApp.Name)
	err = p.appRepo.Upload(config, createdApp)
	if err != nil {
		p.ui.Failed("Error uploading app", err)
		return
	}
	p.ui.Ok()

}
