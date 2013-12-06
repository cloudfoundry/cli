package application

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/formatters"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"os"
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
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c, "push")
		err = errors.New("Incorrect Usage")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd Push) Run(c *cli.Context) {
	app, didCreate := cmd.app(c)
	if !didCreate {
		app = cmd.updateApp(app, c)
	}

	cmd.bindAppToRoute(app, didCreate, c)

	cmd.ui.Say("Uploading %s...", terminal.EntityNameColor(app.Name))

	apiResponse := cmd.appBitsRepo.UploadApp(app.Guid, cmd.path(c))
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.restart(app, didCreate, c)
}

func (cmd Push) bindAppToRoute(app cf.Application, didCreateApp bool, c *cli.Context) {
	if c.Bool("no-route") {
		return
	}

	if len(app.Routes) > 0 && !cmd.routeFlagsPresent(c) {
		return
	}

	if len(app.Routes) == 0 && didCreateApp == false && !cmd.routeFlagsPresent(c) {
		cmd.ui.Say("App %s currently exists as a worker, skipping route creation", terminal.EntityNameColor(app.Name))
		return
	}

	hostName := cmd.hostname(c, app.Name)
	domain := cmd.domain(c)
	route := cmd.route(hostName, domain.DomainFields)

	for _, boundRoute := range app.Routes {
		if boundRoute.Guid == route.Guid {
			return
		}
	}

	cmd.ui.Say("Binding %s to %s...", terminal.EntityNameColor(domain.UrlForHost(hostName)), terminal.EntityNameColor(app.Name))

	apiResponse := cmd.routeRepo.Bind(route.Guid, app.Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")
}

func (cmd Push) restart(app cf.Application, didCreate bool, c *cli.Context) {
	if !didCreate {
		cmd.ui.Say("")
		app, _ = cmd.stopper.ApplicationStop(app)
	}

	cmd.ui.Say("")

	if !c.Bool("no-start") {
		cmd.starter.ApplicationStart(app)
	}
}

func (cmd Push) routeFlagsPresent(c *cli.Context) bool {
	return c.String("n") != "" || c.String("d") != "" || c.Bool("no-hostname")
}

func (cmd Push) route(hostName string, domain cf.DomainFields) (routeFields cf.RouteFields) {
	route, apiResponse := cmd.routeRepo.FindByHostAndDomain(hostName, domain.Name)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Say("Creating route %s...", terminal.EntityNameColor(domain.UrlForHost(hostName)))

		routeFields, apiResponse = cmd.routeRepo.Create(hostName, domain.Guid)
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(apiResponse.Message)
			return
		}

		cmd.ui.Ok()
		cmd.ui.Say("")
	} else {
		cmd.ui.Say("Using route %s", terminal.EntityNameColor(route.URL()))
		routeFields = route.RouteFields
	}

	return
}

func (cmd Push) domain(c *cli.Context) cf.Domain {
	var (
		apiResponse net.ApiResponse
		domain      cf.Domain
		domainName  = c.String("d")
	)

	if domainName != "" {
		domain, apiResponse = cmd.domainRepo.FindByNameInCurrentSpace(domainName)
	} else {
		domain, apiResponse = cmd.domainRepo.FindDefaultAppDomain()
	}

	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
	}

	return domain
}

func (cmd Push) hostname(c *cli.Context, defaultName string) (hostName string) {
	if !c.Bool("no-hostname") {
		hostName = c.String("n")
		if hostName == "" {
			hostName = defaultName
		}
	}
	return
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

func (cmd Push) app(c *cli.Context) (app cf.Application, didCreate bool) {
	appName := c.Args()[0]

	app, apiResponse := cmd.appRepo.Read(appName)
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
	buildpackUrl, stackGuid, command, memoryString, instances := cmd.userAppFields(c)

	if memoryString == "" {
		memoryString = "128M"
	}
	memory := cmd.memoryLimit(memoryString)

	if instances == -1 {
		instances = 1
	}

	cmd.ui.Say("Creating app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(appName),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.SpaceFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	app, apiResponse = cmd.appRepo.Create(appName, buildpackUrl, stackGuid, command, memory, instances)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	return
}

func (cmd Push) updateApp(app cf.Application, c *cli.Context) (updatedApp cf.Application) {
	cmd.ui.Say("Updating app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.SpaceFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	buildpackUrl, stackGuid, command, memoryString, instanceCount := cmd.userAppFields(c)
	params := app.ToParams()

	if buildpackUrl != "" {
		params.Fields["buildpack"] = buildpackUrl
	}
	if command != "" {
		params.Fields["command"] = command
	}
	if instanceCount != -1 {
		params.Fields["instances"] = instanceCount
	}
	if memoryString != "" {
		params.Fields["memory"] = cmd.memoryLimit(memoryString)
	}
	if stackGuid != "" {
		params.Fields["stack_guid"] = stackGuid
	}

	updatedApp, apiResponse := cmd.appRepo.Update(updatedApp.Guid, params)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	return
}

func (cmd Push) userAppFields(c *cli.Context) (buildpackUrl, stackGuid, command, memoryString string, instances int) {
	buildpackUrl = c.String("b")
	memoryString = c.String("m")
	command = c.String("c")
	instances = c.Int("i")

	stackName := c.String("s")
	var (
		stack       cf.Stack
		apiResponse net.ApiResponse
	)
	if stackName != "" {
		stack, apiResponse = cmd.stackRepo.FindByName(stackName)

		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(apiResponse.Message)
			return
		}
		cmd.ui.Say("Using stack %s...", terminal.EntityNameColor(stack.Name))
		stackGuid = stack.Guid
	}

	return
}

func (cmd Push) memoryLimit(s string) (memory uint64) {
	var err error

	memory, err = formatters.MegabytesFromString(s)
	if err != nil {
		cmd.ui.Failed("Invalid memory param: %s\n%s", s, err)
	}

	return
}
