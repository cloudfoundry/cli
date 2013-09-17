package commands

import (
	"cf"
	"cf/api"
	"cf/net"
	"cf/requirements"
	term "cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"strconv"
	"strings"
)

type Push struct {
	ui         term.UI
	starter    ApplicationStarter
	stopper    ApplicationStopper
	zipper     cf.Zipper
	appRepo    api.ApplicationRepository
	domainRepo api.DomainRepository
	routeRepo  api.RouteRepository
	stackRepo  api.StackRepository
}

func NewPush(ui term.UI, starter ApplicationStarter, stopper ApplicationStopper, zipper cf.Zipper,
	aR api.ApplicationRepository, dR api.DomainRepository, rR api.RouteRepository, sR api.StackRepository) (p Push) {
	p.ui = ui
	p.starter = starter
	p.stopper = stopper
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
		reqFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (p Push) Run(c *cli.Context) {
	var err error

	appName := c.String("name")
	app, apiErr := p.appRepo.FindByName(appName)

	if apiErr != nil {
		app, apiErr = p.createApp(appName, c)

		if apiErr != nil {
			return
		}
	}

	p.ui.Say("Uploading %s...", term.EntityNameColor(app.Name))

	dir := c.String("path")
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			p.ui.Failed(err.Error())
			return
		}
	}

	zipBuffer, err := p.zipper.Zip(dir)
	if err != nil {
		p.ui.Failed(err.Error())
		return
	}

	apiErr = p.appRepo.Upload(app, zipBuffer)
	if apiErr != nil {
		p.ui.Failed(apiErr.Error())
		return
	}

	p.ui.Ok()
	p.stopper.ApplicationStop(app)
	if !c.Bool("no-start") {
		p.starter.ApplicationStart(app)
	}
}

func (p Push) createApp(appName string, c *cli.Context) (app cf.Application, apiErr *net.ApiError) {
	newApp := cf.Application{
		Name:         appName,
		Instances:    c.Int("instances"),
		Memory:       getMemoryLimit(c.String("memory")),
		BuildpackUrl: c.String("buildpack"),
	}

	stackName := c.String("stack")
	if stackName != "" {
		var stack cf.Stack
		stack, apiErr = p.stackRepo.FindByName(stackName)

		if apiErr != nil {
			p.ui.Failed(apiErr.Error())
			return
		}
		newApp.Stack = stack
		p.ui.Say("Using stack %s.", term.EntityNameColor(stack.Name))
	}

	p.ui.Say("Creating %s...", term.EntityNameColor(appName))
	app, apiErr = p.appRepo.Create(newApp)
	if apiErr != nil {
		p.ui.Failed(apiErr.Error())
		return
	}
	p.ui.Ok()

	domain, apiErr := p.domainRepo.FindByName(c.String("domain"))

	if apiErr != nil {
		p.ui.Failed(apiErr.Error())
		return
	}

	hostName := c.String("host")
	if hostName == "" {
		hostName = app.Name
	}

	route, apiErr := p.routeRepo.FindByHost(hostName)

	if apiErr != nil {
		newRoute := cf.Route{Host: hostName}

		createdUrl := fmt.Sprintf("%s.%s", newRoute.Host, domain.Name)
		p.ui.Say("Creating route %s...", term.EntityNameColor(createdUrl))
		route, apiErr = p.routeRepo.Create(newRoute, domain)
		if apiErr != nil {
			p.ui.Failed(apiErr.Error())
			return
		}
		p.ui.Ok()
	} else {
		existingUrl := fmt.Sprintf("%s.%s", route.Host, domain.Name)
		p.ui.Say("Using route %s", term.EntityNameColor(existingUrl))
	}

	finalUrl := fmt.Sprintf("%s.%s", route.Host, domain.Name)
	p.ui.Say("Binding %s to %s...", term.EntityNameColor(finalUrl), term.EntityNameColor(app.Name))
	apiErr = p.routeRepo.Bind(route, app)
	if apiErr != nil {
		p.ui.Failed(apiErr.Error())
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
