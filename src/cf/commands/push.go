package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"github.com/codegangsta/cli"
	"strconv"
	"strings"
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
	newApp := cf.Application{
		Name:         appName,
		Instances:    c.Int("instances"),
		Memory:       getMemoryLimit(c.String("memory")),
		BuildpackUrl: c.String("buildpack"),
	}

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

	hostName := c.String("host")
	if hostName == "" {
		hostName = app.Name
	}
	newRoute := cf.Route{Host: hostName}

	p.ui.Say("Creating route %s.%s...", newRoute.Host, domain.Name)
	createdRoute, err := p.routeRepo.Create(config, newRoute, domain)
	if err != nil {
		p.ui.Failed("Error creating route", err)
		return
	}
	p.ui.Ok()

	p.ui.Say("Binding %s.%s to %s...", createdRoute.Host, domain.Name, app.Name)
	err = p.routeRepo.Bind(config, createdRoute, app)
	if err != nil {
		p.ui.Failed("Error binding route", err)
		return
	}
	p.ui.Ok()

	return
}

func getMemoryLimit(arg string) (memory int) {
	var err error

	switch {
	case strings.HasSuffix(arg, "M"):
		trimmedArg := arg[:len(arg)-1]
		memory, err = strconv.Atoi(trimmedArg)
	case strings.HasSuffix(arg, "G"):
		trimmedArg := arg[:len(arg)-1]
		memory, err = strconv.Atoi(trimmedArg)
		memory = memory * 1024
	default:
		memory, err = strconv.Atoi(arg)
	}

	if err != nil {
		memory = 128
	}

	return
}
