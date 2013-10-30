package application

import (
	"cf"
	"cf/api"
	"cf/configuration"
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
	config      *configuration.Configuration
	starter     ApplicationStarter
	stopper     ApplicationStopper
	appRepo     api.ApplicationRepository
	domainRepo  api.DomainRepository
	routeRepo   api.RouteRepository
	stackRepo   api.StackRepository
	appBitsRepo api.ApplicationBitsRepository
}

func NewPush(ui terminal.UI, config *configuration.Configuration, starter ApplicationStarter, stopper ApplicationStopper,
	aR api.ApplicationRepository, dR api.DomainRepository, rR api.RouteRepository, sR api.StackRepository,
	appBitsRepo api.ApplicationBitsRepository) (cmd Push) {

	cmd.ui = ui
	cmd.config = config
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

	app, apiResponse := cmd.appRepo.FindByName(appName)
	if apiResponse.IsError() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	if apiResponse.IsNotFound() {
		app, apiResponse = cmd.createApp(appName, c)

		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(apiResponse.Message)
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

	apiResponse = cmd.appBitsRepo.UploadApp(app, dir)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}
	cmd.ui.Ok()
	cmd.ui.Say("")

	updatedApp, _ := cmd.stopper.ApplicationStop(app)

	cmd.ui.Say("")

	if !c.Bool("no-start") {
		if c.String("b") != "" {
			updatedApp.BuildpackUrl = c.String("b")
		}
		cmd.starter.ApplicationStart(updatedApp)
	}
}

func (cmd Push) createApp(appName string, c *cli.Context) (app cf.Application, apiResponse net.ApiResponse) {
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
		stack, apiResponse = cmd.stackRepo.FindByName(stackName)

		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(apiResponse.Message)
			return
		}
		newApp.Stack = stack
		cmd.ui.Say("Using stack %s...", terminal.EntityNameColor(stack.Name))
	}

	cmd.ui.Say("Creating app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(appName),
		terminal.EntityNameColor(cmd.config.Organization.Name),
		terminal.EntityNameColor(cmd.config.Space.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)
	app, apiResponse = cmd.appRepo.Create(newApp)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}
	cmd.ui.Ok()
	cmd.ui.Say("")

	if !c.Bool("no-route") {
		domainName := c.String("d")

		var hostName string

		if !c.Bool("no-hostname") {
			hostName = c.String("n")
			if hostName == "" {
				hostName = app.Name
			}
		}

		cmd.bindAppToRoute(app, hostName, domainName)
	}

	return
}

func (cmd Push) bindAppToRoute(app cf.Application, hostName, domainName string) {
	var (
		apiResponse net.ApiResponse
		domain      cf.Domain
	)

	if domainName != "" {
		domain, apiResponse = cmd.domainRepo.FindByNameInCurrentSpace(domainName)
	} else {
		domain, apiResponse = cmd.domainRepo.FindDefaultAppDomain()
	}

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	route, apiResponse := cmd.routeRepo.FindByHost(hostName)

	if apiResponse.IsNotSuccessful() {
		newRoute := cf.Route{Host: hostName}

		createdUrl := fmt.Sprintf("%s.%s", newRoute.Host, domain.Name)
		cmd.ui.Say("Creating route %s...", terminal.EntityNameColor(createdUrl))

		route, apiResponse = cmd.routeRepo.Create(newRoute, domain)
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(apiResponse.Message)
			return
		}
		cmd.ui.Ok()
		cmd.ui.Say("")
	} else {
		var existingUrl string

		if (route.Host != "") {
			existingUrl = fmt.Sprintf("%s.%s", route.Host, domain.Name)
		} else {
			existingUrl = domain.Name
		}

		cmd.ui.Say("Using route %s", terminal.EntityNameColor(existingUrl))
	}

	var finalUrl string

	if route.Host != "" {
		finalUrl = fmt.Sprintf("%s.%s", route.Host, domain.Name)
	} else {
		finalUrl = domain.Name
	}

	cmd.ui.Say("Binding %s to %s...", terminal.EntityNameColor(finalUrl), terminal.EntityNameColor(app.Name))
	apiResponse = cmd.routeRepo.Bind(route, app)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")
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
