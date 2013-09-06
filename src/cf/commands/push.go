package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
	"os"
	"strconv"
	"strings"
)

type Push struct {
	ui         term.UI
	config     *configuration.Configuration
	starter    ApplicationStarter
	zipper     cf.Zipper
	appRepo    api.ApplicationRepository
	domainRepo api.DomainRepository
	routeRepo  api.RouteRepository
	stackRepo  api.StackRepository
}

func NewPush(ui term.UI, config *configuration.Configuration, starter ApplicationStarter, zipper cf.Zipper,
	aR api.ApplicationRepository, dR api.DomainRepository, rR api.RouteRepository, sR api.StackRepository) (p Push) {
	p.ui = ui
	p.config = config
	p.starter = starter
	p.zipper = zipper
	p.appRepo = aR
	p.domainRepo = dR
	p.routeRepo = rR
	p.stackRepo = sR
	return
}

func (cmd Push) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewSpaceRequirement(),
	}
	return
}

func (p Push) Run(c *cli.Context) {
	appName := c.String("name")
	app, err := p.appRepo.FindByName(appName)

	if err != nil {
		app, err = p.createApp(p.config, appName, c)

		if err != nil {
			return
		}
	}

	p.ui.Say("Uploading %s...", app.Name)

	dir := c.String("path")
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			p.ui.Failed("Error getting working directory", err)
			return
		}
	}

	zipBuffer, err := p.zipper.Zip(dir)
	if err != nil {
		p.ui.Failed("Error zipping app", err)
		return
	}

	err = p.appRepo.Upload(p.config, app, zipBuffer)
	if err != nil {
		p.ui.Failed("Error uploading app", err)
		return
	}

	p.ui.Ok()
	if !c.Bool("no-start") {
		p.starter.ApplicationStart(app)
	}
}

func (p Push) createApp(config *configuration.Configuration, appName string, c *cli.Context) (app cf.Application, err error) {
	newApp := cf.Application{
		Name:         appName,
		Instances:    c.Int("instances"),
		Memory:       getMemoryLimit(c.String("memory")),
		BuildpackUrl: c.String("buildpack"),
	}

	stackName := c.String("stack")
	if stackName != "" {
		var stack cf.Stack
		stack, err = p.stackRepo.FindByName(p.config, stackName)

		if err != nil {
			p.ui.Failed("Error finding stack", err)
			return
		}
		newApp.Stack = stack
		p.ui.Say("Using stack %s.", stack.Name)
	}

	p.ui.Say("Creating %s...", appName)
	app, err = p.appRepo.Create(newApp)
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

	route, err := p.routeRepo.FindByHost(config, hostName)
	if err != nil {
		newRoute := cf.Route{Host: hostName}

		p.ui.Say("Creating route %s.%s...", newRoute.Host, domain.Name)
		route, err = p.routeRepo.Create(config, newRoute, domain)
		if err != nil {
			p.ui.Failed("Error creating route", err)
			return
		}
		p.ui.Ok()
	} else {
		p.ui.Say("Using route %s.%s", route.Host, domain.Name)
	}

	p.ui.Say("Binding %s.%s to %s...", route.Host, domain.Name, app.Name)
	err = p.routeRepo.Bind(config, route, app)
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
