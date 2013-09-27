package commands

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

	app, found, apiErr := cmd.appRepo.FindByName(appName)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	if !found {
		app, apiErr = cmd.createApp(appName, c)

		if apiErr != nil {
			cmd.ui.Failed(apiErr.Error())
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

	apiErr = cmd.appBitsRepo.UploadApp(app, dir)

	cmd.ui.Ok()
	updatedApp, _ := cmd.stopper.ApplicationStop(app)
	if !c.Bool("no-start") {
		if c.String("b") != "" {
			updatedApp.BuildpackUrl = c.String("b")
		}
		cmd.starter.ApplicationStart(updatedApp)
	}
}

func (cmd Push) createApp(appName string, c *cli.Context) (app cf.Application, apiErr *net.ApiError) {
	newApp := cf.Application{
		Name:         appName,
		Instances:    c.Int("i"),
		Memory:       getMemoryLimit(c.String("m")),
		BuildpackUrl: c.String("b"),
	}

	stackName := c.String("s")
	if stackName != "" {
		var stack cf.Stack
		stack, apiErr = cmd.stackRepo.FindByName(stackName)

		if apiErr != nil {
			cmd.ui.Failed(apiErr.Error())
			return
		}
		newApp.Stack = stack
		cmd.ui.Say("Using stack %s.", terminal.EntityNameColor(stack.Name))
	}

	cmd.ui.Say("Creating %s...", terminal.EntityNameColor(appName))
	app, apiErr = cmd.appRepo.Create(newApp)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
	cmd.ui.Ok()

	domain, apiErr := cmd.domainRepo.FindByName(c.String("d"))

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	hostName := c.String("n")
	if hostName == "" {
		hostName = app.Name
	}

	route, apiErr := cmd.routeRepo.FindByHost(hostName)

	if apiErr != nil {
		newRoute := cf.Route{Host: hostName}

		createdUrl := fmt.Sprintf("%s.%s", newRoute.Host, domain.Name)
		cmd.ui.Say("Creating route %s...", terminal.EntityNameColor(createdUrl))
		route, apiErr = cmd.routeRepo.Create(newRoute, domain)
		if apiErr != nil {
			cmd.ui.Failed(apiErr.Error())
			return
		}
		cmd.ui.Ok()
	} else {
		existingUrl := fmt.Sprintf("%s.%s", route.Host, domain.Name)
		cmd.ui.Say("Using route %s", terminal.EntityNameColor(existingUrl))
	}

	finalUrl := fmt.Sprintf("%s.%s", route.Host, domain.Name)
	cmd.ui.Say("Binding %s to %s...", terminal.EntityNameColor(finalUrl), terminal.EntityNameColor(app.Name))
	apiErr = cmd.routeRepo.Bind(route, app)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
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
