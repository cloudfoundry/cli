package application

import (
	"cf"
	"cf/api"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"strconv"
	"strings"
)

type Push struct {
	ui          terminal.UI
	starter     ApplicationStarter
	stopper     ApplicationStopper
	appRepo     api.ApplicationRepository
	domainRepo  api.DomainRepository
	routeRepo   api.RouteRepository
	stackRepo   api.StackRepository
	appBitsRepo api.ApplicationBitsRepository
}

func NewPush(ui terminal.UI, starter ApplicationStarter, stopper ApplicationStopper,
	aR api.ApplicationRepository, dR api.DomainRepository, rR api.RouteRepository, sR api.StackRepository,
	appBitsRepo api.ApplicationBitsRepository) (cmd Push) {

	cmd.ui = ui
	cmd.starter = starter
	cmd.stopper = stopper
	cmd.appRepo = aR
	cmd.domainRepo = dR
	cmd.routeRepo = rR
	cmd.stackRepo = sR
	cmd.appBitsRepo = appBitsRepo
	return
}

func (cmd Push) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd Push) Run(c *cli.Context) {
	var err error

	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c, "push")
		return
	}

	appName := c.Args()[0]

	app, apiStatus := cmd.appRepo.FindByName(appName)
	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	if apiStatus.IsNotFound() {
		app, apiStatus = cmd.createApp(appName, c)

		if apiStatus.IsError() {
			cmd.ui.Failed(apiStatus.Message)
			return
		}
	}

	cmd.ui.Say("Uploading %s...", terminal.EntityNameColor(app.Name))

	dir := c.String("p")
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			cmd.ui.Failed(err.Error())
			return
		}
	}

	apiStatus = cmd.appBitsRepo.UploadApp(app, dir)

	cmd.ui.Ok()
	updatedApp, _ := cmd.stopper.ApplicationStop(app)
	if !c.Bool("no-start") {
		if c.String("b") != "" {
			updatedApp.BuildpackUrl = c.String("b")
		}
		cmd.starter.ApplicationStart(updatedApp)
	}
}

func (cmd Push) createApp(appName string, c *cli.Context) (app cf.Application, apiStatus net.ApiStatus) {
	newApp := cf.Application{
		Name:         appName,
		Instances:    c.Int("i"),
		Memory:       getMemoryLimit(c.String("m")),
		BuildpackUrl: c.String("b"),
		Command:      c.String("c"),
	}

	stackName := c.String("s")
	if stackName != "" {
		var stack cf.Stack
		stack, apiStatus = cmd.stackRepo.FindByName(stackName)

		if apiStatus.IsError() {
			cmd.ui.Failed(apiStatus.Message)
			return
		}
		newApp.Stack = stack
		cmd.ui.Say("Using stack %s.", terminal.EntityNameColor(stack.Name))
	}

	cmd.ui.Say("Creating %s...", terminal.EntityNameColor(appName))
	app, apiStatus = cmd.appRepo.Create(newApp)
	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}
	cmd.ui.Ok()

	domain, apiStatus := cmd.domainRepo.FindByNameInCurrentSpace(c.String("d"))

	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	hostName := c.String("n")
	if hostName == "" {
		hostName = app.Name
	}

	route, apiStatus := cmd.routeRepo.FindByHost(hostName)

	if apiStatus.IsError() {
		newRoute := cf.Route{Host: hostName}

		createdUrl := fmt.Sprintf("%s.%s", newRoute.Host, domain.Name)
		cmd.ui.Say("Creating route %s...", terminal.EntityNameColor(createdUrl))
		route, apiStatus = cmd.routeRepo.Create(newRoute, domain)
		if apiStatus.IsError() {
			cmd.ui.Failed(apiStatus.Message)
			return
		}
		cmd.ui.Ok()
	} else {
		existingUrl := fmt.Sprintf("%s.%s", route.Host, domain.Name)
		cmd.ui.Say("Using route %s", terminal.EntityNameColor(existingUrl))
	}

	finalUrl := fmt.Sprintf("%s.%s", route.Host, domain.Name)
	cmd.ui.Say("Binding %s to %s...", terminal.EntityNameColor(finalUrl), terminal.EntityNameColor(app.Name))
	apiStatus = cmd.routeRepo.Bind(route, app)
	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}
	cmd.ui.Ok()

	return
}

func getMemoryLimit(arg string) (memory uint64) {
	var err error

	switch {
	case strings.HasSuffix(arg, "M"):
		trimmedArg := arg[:len(arg)-1]
		memory, err = strconv.ParseUint(trimmedArg, 10, 0)
	case strings.HasSuffix(arg, "G"):
		trimmedArg := arg[:len(arg)-1]
		memory, err = strconv.ParseUint(trimmedArg, 10, 0)
		memory = memory * 1024
	default:
		memory, err = strconv.ParseUint(arg, 10, 0)
	}

	if err != nil {
		memory = 128
	}

	return
}
