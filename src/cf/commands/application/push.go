package application

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
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
	var (
		apiResponse net.ApiResponse
	)

	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c, "push")
		return
	}

	app, didCreate := cmd.getApp(c)

	domain := cmd.domain(c)
	hostName := cmd.hostName(app, c)
	cmd.bindAppToRoute(app, domain, hostName, didCreate, c)

	cmd.ui.Say("Uploading %s...", terminal.EntityNameColor(app.Name))

	dir := cmd.path(c)
	apiResponse = cmd.appBitsRepo.UploadApp(app.Guid, dir)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	cmd.restart(app, c)
}

func (cmd Push) getApp(c *cli.Context) (app cf.Application, didCreate bool) {
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
		didCreate = true
	}

	return
}

func (cmd Push) createApp(appName string, c *cli.Context) (app cf.Application, apiResponse net.ApiResponse) {
	buildpackUrl := c.String("b")
	instances := c.Int("i")
	memory := memoryLimit(c.String("m"))
	command := c.String("c")
	stackName := c.String("s")

	var stack cf.Stack
	if stackName != "" {
		stack, apiResponse = cmd.stackRepo.FindByName(stackName)

		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(apiResponse.Message)
			return
		}
		cmd.ui.Say("Using stack %s...", terminal.EntityNameColor(stack.Name))
	}

	cmd.ui.Say("Creating app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(appName),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.SpaceFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	app, apiResponse = cmd.appRepo.Create(appName, buildpackUrl, stack.Guid, command, memory, instances)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	return
}

func (cmd Push) domain(c *cli.Context) (domain cf.Domain) {
	var apiResponse net.ApiResponse

	domainName := c.String("d")

	if domainName != "" {
		domain, apiResponse = cmd.domainRepo.FindByNameInCurrentSpace(domainName)
	} else {
		domain, apiResponse = cmd.domainRepo.FindDefaultAppDomain()
	}

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
	}
	return
}

func (cmd Push) hostName(app cf.Application, c *cli.Context) (hostName string) {
	if !c.Bool("no-hostname") {
		hostName = c.String("n")
		if hostName == "" {
			hostName = app.Name
		}
	}
	return
}

func (cmd Push) createRoute(hostName string, domain cf.Domain) (route cf.RouteFields) {
	cmd.ui.Say("Creating route %s...", terminal.EntityNameColor(domain.UrlForHost(hostName)))

	route, apiResponse := cmd.routeRepo.Create(hostName, domain.Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	return
}

func (cmd Push) bindAppToRoute(app cf.Application, domain cf.Domain, hostName string, didCreate bool, c *cli.Context) {
	if c.Bool("no-route") {
		return
	}

	if len(app.Routes) == 0 && didCreate == false {
		cmd.ui.Say("App %s currently exists as a worker, skipping route creation", terminal.EntityNameColor(app.Name))
		return
	}

	routeGuid := ""
	route, apiResponse := cmd.routeRepo.FindByHostAndDomain(hostName, domain.Name)
	if apiResponse.IsNotSuccessful() {
		routeGuid = cmd.createRoute(hostName, domain).Guid
	} else {
		routeGuid = route.Guid
		cmd.ui.Say("Using route %s", terminal.EntityNameColor(route.URL()))
	}

	for _, boundRoute := range app.Routes {
		if boundRoute.Guid == routeGuid {
			return
		}
	}

	cmd.ui.Say("Binding %s to %s...", terminal.EntityNameColor(domain.UrlForHost(hostName)), terminal.EntityNameColor(app.Name))

	apiResponse = cmd.routeRepo.Bind(routeGuid, app.Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")
}

func (cmd Push) path(c *cli.Context) (dir string) {
	dir = c.String("p")
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			cmd.ui.Failed(err.Error())
			return
		}
	}
	return
}

func (cmd Push) restart(app cf.Application, c *cli.Context) {
	updatedApp, _ := cmd.stopper.ApplicationStop(app)

	cmd.ui.Say("")

	if !c.Bool("no-start") {
		if buildpackUrl := c.String("b"); buildpackUrl == "" {
			cmd.starter.ApplicationStart(updatedApp)
		} else {
			cmd.starter.ApplicationStartWithBuildpack(updatedApp, buildpackUrl)
		}
	}
}

func memoryLimit(arg string) (memory uint64) {
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
