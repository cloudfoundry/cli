package application

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/manifest"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"generic"
	"github.com/codegangsta/cli"
	"os"
)

type Push struct {
	ui           terminal.UI
	appParams    cf.AppParams
	config       *configuration.Configuration
	manifestRepo manifest.ManifestRepository
	starter      ApplicationStarter
	stopper      ApplicationStopper
	appRepo      api.ApplicationRepository
	domainRepo   api.DomainRepository
	routeRepo    api.RouteRepository
	stackRepo    api.StackRepository
	appBitsRepo  api.ApplicationBitsRepository
}

func NewPush(ui terminal.UI, config *configuration.Configuration, manifestRepo manifest.ManifestRepository, starter ApplicationStarter, stopper ApplicationStopper,
	aR api.ApplicationRepository, dR api.DomainRepository, rR api.RouteRepository, sR api.StackRepository,
	appBitsRepo api.ApplicationBitsRepository) (cmd *Push) {
	cmd = &Push{}
	cmd.ui = ui
	cmd.config = config
	cmd.manifestRepo = manifestRepo
	cmd.starter = starter
	cmd.stopper = stopper
	cmd.appRepo = aR
	cmd.domainRepo = dR
	cmd.routeRepo = rR
	cmd.stackRepo = sR
	cmd.appBitsRepo = appBitsRepo
	return
}

func (cmd *Push) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	path := cmd.path(c)
	m, err := cmd.manifestRepo.ReadManifest(path)
	if err != nil {
		cmd.ui.Failed("Error reading manifest from path:%s/n%s", path, err)
		return
	}

	manifestParams := cf.NewEmptyAppParams()
	if len(m.Applications) > 0 {
		manifestParams = m.Applications[0]
	}

	contextParams, err := cf.NewAppParamsFromContext(c)
	if err != nil {
		cmd.ui.Failed("Error: %s", err)
		return
	}

	appFields := cf.NewAppParams(generic.Merge(manifestParams, contextParams))

	if !appFields.Has("name") {
		cmd.ui.FailWithUsage(c, "push")
		err = errors.New("Incorrect Usage")
		return
	}

	cmd.appParams = appFields

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd *Push) Run(c *cli.Context) {
	cmd.fetchStackGuid()

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
	cmd.restart(app, c)
}

func (cmd *Push) fetchStackGuid() {
	if !cmd.appParams.Has("stack") {
		return
	}
	stackName := cmd.appParams.Get("stack").(string)

	cmd.ui.Say("Using stack %s...", terminal.EntityNameColor(stackName))

	stack, apiResponse := cmd.stackRepo.FindByName(stackName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}
	cmd.ui.Ok()

	cmd.appParams.Set("stack_guid", stack.Guid)
}

func (cmd *Push) bindAppToRoute(app cf.Application, didCreateApp bool, c *cli.Context) {
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

func (cmd *Push) restart(app cf.Application, c *cli.Context) {
	if app.State != "stopped" {
		cmd.ui.Say("")
		app, _ = cmd.stopper.ApplicationStop(app)
	}

	cmd.ui.Say("")

	if !c.Bool("no-start") {
		cmd.starter.ApplicationStart(app)
	}
}

func (cmd *Push) routeFlagsPresent(c *cli.Context) bool {
	return c.String("n") != "" || c.String("d") != "" || c.Bool("no-hostname")
}

func (cmd *Push) route(hostName string, domain cf.DomainFields) (route cf.Route) {
	route, apiResponse := cmd.routeRepo.FindByHostAndDomain(hostName, domain.Name)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Say("Creating route %s...", terminal.EntityNameColor(domain.UrlForHost(hostName)))

		route, apiResponse = cmd.routeRepo.Create(hostName, domain.Guid)
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(apiResponse.Message)
			return
		}

		cmd.ui.Ok()
		cmd.ui.Say("")
	} else {
		cmd.ui.Say("Using route %s", terminal.EntityNameColor(route.URL()))
	}

	return
}

func (cmd *Push) domain(c *cli.Context) (domain cf.Domain) {
	var (
		apiResponse net.ApiResponse
		domainName  = c.String("d")
	)

	if domainName != "" {
		domain, apiResponse = cmd.domainRepo.FindByNameInCurrentSpace(domainName)
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(apiResponse.Message)
		}
	} else {
		stopChan := make(chan bool, 1)
		domainsChan, statusChan := cmd.domainRepo.ListDomainsForOrg(cmd.config.OrganizationFields.Guid, stopChan)

		for domainsChunk := range domainsChan {
			for _, d := range domainsChunk {
				if d.Shared {
					domain = d
					stopChan <- true
					break
				}
			}
		}

		apiResponse, ok := <-statusChan
		if (domain.Guid == "") && ok && apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(apiResponse.Message)
		} else if domain.Guid == "" {
			cmd.ui.Failed("No default domain exists")
		}
	}

	return
}

func (cmd *Push) hostname(c *cli.Context, defaultName string) (hostName string) {
	if !c.Bool("no-hostname") {
		hostName = c.String("n")
		if hostName == "" {
			hostName = defaultName
		}
	}
	return
}

func (cmd *Push) path(c *cli.Context) (dir string) {
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

func (cmd *Push) app(c *cli.Context) (app cf.Application, didCreate bool) {
	app, apiResponse := cmd.appRepo.Read(cmd.appParams.Get("name").(string))
	if apiResponse.IsError() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	if apiResponse.IsNotFound() {
		app, apiResponse = cmd.createApp(c)
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(apiResponse.Message)
			return
		}
		didCreate = true
	}

	return
}

func (cmd *Push) createApp(c *cli.Context) (app cf.Application, apiResponse net.ApiResponse) {
	cmd.appParams.Set("space_guid", cmd.config.SpaceFields.Guid)

	cmd.ui.Say("Creating app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(cmd.appParams.Get("name").(string)),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.SpaceFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	app, apiResponse = cmd.appRepo.Create(cmd.appParams)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	return
}

func (cmd *Push) updateApp(app cf.Application, c *cli.Context) (updatedApp cf.Application) {
	cmd.ui.Say("Updating app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.SpaceFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	updatedApp, apiResponse := cmd.appRepo.Update(app.Guid, cmd.appParams)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	return
}
