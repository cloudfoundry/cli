package application

import (
	"cf"
	"cf/api"
	"cf/commands/service"
	"cf/configuration"
	"cf/manifest"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"generic"
	"github.com/codegangsta/cli"
	"os"
	"path/filepath"
)

type Push struct {
	ui             terminal.UI
	appSet         cf.AppSet
	config         *configuration.Configuration
	manifestRepo   manifest.ManifestRepository
	starter        ApplicationStarter
	stopper        ApplicationStopper
	binder         service.ServiceBinder
	appRepo        api.ApplicationRepository
	domainRepo     api.DomainRepository
	routeRepo      api.RouteRepository
	serviceRepo    api.ServiceRepository
	stackRepo      api.StackRepository
	appBitsRepo    api.ApplicationBitsRepository
	globalServices cf.ServiceInstanceSet
}

func NewPush(ui terminal.UI, config *configuration.Configuration, manifestRepo manifest.ManifestRepository,
	starter ApplicationStarter, stopper ApplicationStopper, binder service.ServiceBinder,
	appRepo api.ApplicationRepository, domainRepo api.DomainRepository, routeRepo api.RouteRepository,
	stackRepo api.StackRepository, serviceRepo api.ServiceRepository, appBitsRepo api.ApplicationBitsRepository) (cmd *Push) {
	cmd = &Push{}
	cmd.ui = ui
	cmd.config = config
	cmd.manifestRepo = manifestRepo
	cmd.starter = starter
	cmd.stopper = stopper
	cmd.binder = binder
	cmd.appRepo = appRepo
	cmd.domainRepo = domainRepo
	cmd.routeRepo = routeRepo
	cmd.serviceRepo = serviceRepo
	cmd.stackRepo = stackRepo
	cmd.appBitsRepo = appBitsRepo
	return
}

func createAppSetFromContextAndManifest(
	contextParams cf.AppParams,
	rootAppPath string,
	m *manifest.Manifest) (appSet cf.AppSet, err error) {
	if len(m.Applications) == 0 {
		appSet = cf.NewAppSet(contextParams)
	} else {
		appSet = cf.NewEmptyAppSet()

		for _, manifestAppParams := range m.Applications {
			appFields := cf.NewAppParams(generic.Merge(manifestAppParams, contextParams))

			path := rootAppPath
			if manifestAppParams.Has("path") {
				path = filepath.Join(rootAppPath, manifestAppParams.Get("path").(string))
			}
			appFields.Set("path", path)

			appSet = append(appSet, appFields)
		}
	}

	for _, appParams := range appSet {
		if !appParams.Has("name") {
			err = errors.New("app name is a required field")
		}
	}

	return
}

func (cmd *Push) appAndManifestPaths(userSpecifiedAppPath, userSpecifiedManifestPath string) (appPath, manifestPath string) {
	cwd, _ := os.Getwd()

	if userSpecifiedAppPath != "" && userSpecifiedManifestPath != "" {
		cmd.ui.Warn("-p is ignored when using a manifest. Please specify the path in the manifest.")
	}

	if userSpecifiedAppPath != "" {
		appPath = userSpecifiedAppPath
	} else if userSpecifiedManifestPath != "" {
		appPath = userSpecifiedManifestPath
	} else {
		appPath = cwd
	}

	if userSpecifiedManifestPath != "" {
		manifestPath = userSpecifiedManifestPath
	} else if userSpecifiedAppPath != "" {
		manifestPath = userSpecifiedAppPath
	} else {
		manifestPath = cwd
	}

	return
}

func (cmd *Push) findAndValidateAppsToPush(c *cli.Context) {
	appPath, manifestPath := cmd.appAndManifestPaths(c.String("p"), c.String("manifest"))

	manifest, errs := cmd.manifestRepo.ReadManifest(manifestPath)
	if !errs.Empty() {
		cmd.ui.Failed("Error reading manifest file: \n%s", errs)
		return
	}

	appParams, err := cf.NewAppParamsFromContext(c)
	if err != nil {
		cmd.ui.Failed("Error: %s", err)
		return
	}

	appParams.Set("path", appPath)

	cmd.appSet, err = createAppSetFromContextAndManifest(appParams, appPath, manifest)
	if err != nil {
		cmd.ui.Failed("Error: %s", err)
		return
	}
}

func (cmd *Push) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd *Push) Run(c *cli.Context) {
	cmd.findAndValidateAppsToPush(c)

	for _, appParams := range cmd.appSet {
		cmd.fetchStackGuid(&appParams)

		app, didCreate := cmd.app(appParams)
		if !didCreate {
			app = cmd.updateApp(app, appParams)
		}

		cmd.bindAppToRoute(app, appParams, didCreate, c)

		cmd.ui.Say("Uploading %s...", terminal.EntityNameColor(app.Name))

		apiResponse := cmd.appBitsRepo.UploadApp(app.Guid, appParams.Get("path").(string))
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(apiResponse.Message)
			return
		}
		cmd.ui.Ok()

		if appParams.Has("services") {
			services := appParams.Get("services").([]string)

			for _, serviceName := range services {
				serviceInstance, response := cmd.serviceRepo.FindInstanceByName(serviceName)

				if response.IsNotSuccessful() {
					cmd.ui.Failed("Could not find service %s to bind to %s", serviceName, appParams.Get("name").(string))
					return
				}

				cmd.ui.Say("Binding service %s to %s in org %s / space %s as %s", serviceName, appParams.Get("name").(string), cmd.config.OrganizationFields.Name, cmd.config.SpaceFields.Name, cmd.config.Username())
				bindResponse := cmd.binder.BindApplication(app, serviceInstance)
				cmd.ui.Ok()

				if bindResponse.IsNotSuccessful() && bindResponse.ErrorCode != service.AppAlreadyBoundErrorCode {
					cmd.ui.Failed("Could not find to service %s\nError: %s", serviceName, bindResponse.Message)
					return
				}
			}
		}

		cmd.restart(app, appParams, c)
	}
}

func (cmd *Push) fetchStackGuid(appParams *cf.AppParams) {
	if !appParams.Has("stack") {
		return
	}

	stackName := appParams.Get("stack").(string)
	cmd.ui.Say("Using stack %s...", terminal.EntityNameColor(stackName))

	stack, apiResponse := cmd.stackRepo.FindByName(stackName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	appParams.Set("stack_guid", stack.Guid)
}

func (cmd *Push) bindAppToRoute(app cf.Application, params cf.AppParams, didCreateApp bool, c *cli.Context) {
	if c.Bool("no-route") {
		return
	}

	routeFlagsPresent := c.String("n") != "" || c.String("d") != "" || c.Bool("no-hostname")
	if len(app.Routes) > 0 && !routeFlagsPresent {
		return
	}

	if len(app.Routes) == 0 && didCreateApp == false && !routeFlagsPresent {
		cmd.ui.Say("App %s currently exists as a worker, skipping route creation", terminal.EntityNameColor(app.Name))
		return
	}

	var defaultHostname string
	if params.Has("host") {
		defaultHostname = params.Get("host").(string)
	} else {
		defaultHostname = app.Name
	}

	var domainName string
	if params.Has("domain") {
		domainName = params.Get("domain").(string)
	} else {
		domainName = c.String("d")
	}

	hostName := cmd.hostname(c, defaultHostname)
	domain := cmd.domain(c, domainName)
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

func (cmd *Push) restart(app cf.Application, params cf.AppParams, c *cli.Context) {
	if app.State != "stopped" {
		cmd.ui.Say("")
		app, _ = cmd.stopper.ApplicationStop(app)
	}

	cmd.ui.Say("")

	if c.Bool("no-start") {
		return
	}

	if params.Has("health_check_timeout") {
		timeout := params.Get("health_check_timeout").(int)
		cmd.starter.SetStartTimeoutSeconds(timeout)
	}

	cmd.starter.ApplicationStart(app)
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

func (cmd *Push) domain(c *cli.Context, domainName string) (domain cf.Domain) {
	var apiResponse net.ApiResponse

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
	if c.Bool("no-hostname") {
		return
	}

	hostName = c.String("n")
	if hostName == "" {
		hostName = defaultName
	}

	return
}

func (cmd *Push) app(appParams cf.AppParams) (app cf.Application, didCreate bool) {
	appName := appParams.Get("name").(string)
	app, apiResponse := cmd.appRepo.Read(appName)
	if apiResponse.IsError() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	if apiResponse.IsNotFound() {
		app, apiResponse = cmd.createApp(appParams)
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(apiResponse.Message)
			return
		}
		didCreate = true
	}

	return
}

func (cmd *Push) createApp(appParams cf.AppParams) (app cf.Application, apiResponse net.ApiResponse) {
	appParams.Set("space_guid", cmd.config.SpaceFields.Guid)

	cmd.ui.Say("Creating app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(appParams.Get("name").(string)),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.SpaceFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	app, apiResponse = cmd.appRepo.Create(appParams)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	return
}

func (cmd *Push) updateApp(app cf.Application, appParams cf.AppParams) (updatedApp cf.Application) {
	cmd.ui.Say("Updating app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields.Name),
		terminal.EntityNameColor(cmd.config.SpaceFields.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	updatedApp, apiResponse := cmd.appRepo.Update(app.Guid, appParams)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	return
}
