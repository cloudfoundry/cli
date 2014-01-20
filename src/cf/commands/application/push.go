package application

import (
	"cf"
	"cf/api"
	"cf/commands/service"
	"cf/configuration"
	"cf/formatters"
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

func (cmd *Push) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd *Push) Run(c *cli.Context) {
	appSet := cmd.findAndValidateAppsToPush(c)

	for _, appParams := range appSet {
		cmd.fetchStackGuid(appParams)

		app, didCreate := cmd.app(appParams)
		if !didCreate {
			app = cmd.updateApp(app, appParams)
		}

		cmd.bindAppToRoute(app, appParams, didCreate, c)

		cmd.ui.Say("Uploading %s...", terminal.EntityNameColor(app.Name))

		apiResponse := cmd.appBitsRepo.UploadApp(app.Guid, appParams.Get("path").(string), cmd.describeUploadOperation)
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

func (cmd *Push) describeUploadOperation(zipFileBytes, fileCount uint64) {
	humanReadableBytes := formatters.ByteSize(zipFileBytes)
	cmd.ui.Say("Uploading app: %s, %d files", humanReadableBytes, fileCount)
}

func (cmd *Push) fetchStackGuid(appParams cf.AppParams) {
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
	if !appParams.Has("name") {
		cmd.ui.Failed("Error: No name found for app")
		return
	}

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

func (cmd *Push) findAndValidateAppsToPush(c *cli.Context) (appSet cf.AppSet) {
	var m *manifest.Manifest
	if c.String("f") != "" {
		baseManifestPath, manifestFilename := cmd.manifestPathFromContext(c)
		m = cmd.instantiateManifest(c, filepath.Join(baseManifestPath, manifestFilename))
	} else {
		m = manifest.NewEmptyManifest()
	}

	appParams, err := cf.NewAppParamsFromContext(c)
	if err != nil {
		cmd.ui.Failed("Error: %s", err)
		return
	}

	baseAppPath := cmd.appPathFromContext(c)
	appParams.Set("path", baseAppPath)

	appSet, err = createAppSetFromContextAndManifest(appParams, baseAppPath, m)
	if err != nil {
		cmd.ui.Failed("Error: %s", err)
	}
	return
}

func (cmd *Push) appPathFromContext(c *cli.Context) (appPath string) {
	if c.String("p") != "" {
		path, err := filepath.Abs(c.String("p"))
		if err != nil {
			cmd.ui.Failed("Error finding app path: %s", err)
			return
		}

		fileInfo, err := os.Stat(path)
		if err != nil {
			cmd.ui.Failed("Error finding app path: %s", err)
			return
		}

		if fileInfo.IsDir() {
			appPath = path
		} else {
			appPath = filepath.Dir(path)
		}
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			cmd.ui.Failed("Error reading current working directory: %s", err)
			return
		}

		appPath = cwd
	}

	return
}

func (cmd *Push) manifestPathFromContext(c *cli.Context) (basePath, manifestFilename string) {
	var err error

	if c.String("f") != "" {
		basePath, manifestFilename, err = cmd.manifestRepo.ManifestPath(c.String("f"))
	}

	if err != nil {
		cmd.ui.Failed("%s", err)
		return
	}

	return
}

func (cmd *Push) instantiateManifest(c *cli.Context, manifestPath string) (m *manifest.Manifest) {
	if c.Bool("no-manifest") {
		m = manifest.NewEmptyManifest()
		return
	}

	m, errs := cmd.manifestRepo.ReadManifest(manifestPath)

	if !errs.Empty() {
		if os.IsNotExist(errs[0]) && c.String("f") == "" {
			m = manifest.NewEmptyManifest()
			return
		} else {
			cmd.ui.Failed("Error reading manifest file:\n%s", errs)
			return
		}
	}

	cmd.ui.Say("Using manifest file %s\n", terminal.EntityNameColor(manifestPath))
	return
}

func createAppSetFromContextAndManifest(contextParams cf.AppParams, rootAppPath string, m *manifest.Manifest) (appSet cf.AppSet, err error) {
	if contextParams.Has("name") && len(m.Applications) > 1 {
		err = errors.New("APP_NAME command line argument is not allowed when pushing multiple apps from a manifest file.")
		return
	}

	appSet = make([]cf.AppParams, 0, len(m.Applications))
	if len(m.Applications) == 0 {
		appSet = append(appSet, contextParams)
	} else {
		for _, manifestAppParams := range m.Applications {
			appFields := cf.NewAppParams(generic.Merge(manifestAppParams, contextParams))

			path := rootAppPath
			if manifestAppParams.Has("path") {
				pathFromManifest := manifestAppParams.Get("path").(string)
				if filepath.IsAbs(pathFromManifest) {
					path = filepath.Clean(pathFromManifest)
				} else {
					path = filepath.Join(rootAppPath, pathFromManifest)
				}
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
