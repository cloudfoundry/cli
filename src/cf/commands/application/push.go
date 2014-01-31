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
	"fmt"
	"generic"
	"github.com/codegangsta/cli"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

		app := cmd.createOrUpdateApp(appParams)

		cmd.bindAppToRoute(app, appParams, c)

		cmd.ui.Say("Uploading %s...", terminal.EntityNameColor(app.Name))

		apiResponse := cmd.appBitsRepo.UploadApp(app.Guid, appParams.Get("path").(string), cmd.describeUploadOperation)
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(fmt.Sprintf("Error uploading application.\n%s", apiResponse.Message))
			return
		}
		cmd.ui.Ok()

		if appParams.Has("services") {
			cmd.bindAppToServices(appParams.Get("services").([]string), app)
		}

		cmd.restart(app, appParams, c)
	}
}

func (cmd *Push) bindAppToServices(services []string, app cf.Application) {
	for _, serviceName := range services {
		serviceInstance, response := cmd.serviceRepo.FindInstanceByName(serviceName)

		if response.IsNotSuccessful() {
			cmd.ui.Failed("Could not find service %s to bind to %s", serviceName, app.Name)
			return
		}

		cmd.ui.Say("Binding service %s to %s in org %s / space %s as %s", serviceName, app.Name, cmd.config.OrganizationFields.Name, cmd.config.SpaceFields.Name, cmd.config.Username())
		bindResponse := cmd.binder.BindApplication(app, serviceInstance)
		cmd.ui.Ok()

		if bindResponse.IsNotSuccessful() && bindResponse.ErrorCode != service.AppAlreadyBoundErrorCode {
			cmd.ui.Failed("Could not find to service %s\nError: %s", serviceName, bindResponse.Message)
			return
		}
	}
}

func (cmd *Push) describeUploadOperation(path string, zipFileBytes, fileCount uint64) {
	humanReadableBytes := formatters.ByteSize(zipFileBytes)
	cmd.ui.Say("Uploading from: %s\n%s, %d files", path, humanReadableBytes, fileCount)
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

func (cmd *Push) bindAppToRoute(app cf.Application, params cf.AppParams, c *cli.Context) {
	if c.Bool("no-route") {
		return
	}

	if params.Has("no-route") && params.Get("no-route") == true {
		cmd.ui.Say("App %s is a worker, skipping route creation", terminal.EntityNameColor(app.Name))
		return
	}

	routeFlagsPresent := c.String("n") != "" || c.String("d") != "" || c.Bool("no-hostname")
	if len(app.Routes) > 0 && !routeFlagsPresent {
		return
	}

	var defaultHostname string
	if params.Has("host") {
		defaultHostname = params.Get("host").(string)
	} else {
		defaultHostname = hostNameForString(app.Name)
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

var forbiddenHostCharRegex = regexp.MustCompile("[^a-z0-9-]")
var whitespaceRegex = regexp.MustCompile(`[\s_]+`)

func hostNameForString(name string) string {
	nameBytes := []byte(strings.ToLower(name))
	nameBytes = whitespaceRegex.ReplaceAll(nameBytes, []byte("-"))
	nameBytes = forbiddenHostCharRegex.ReplaceAll(nameBytes, []byte{})
	return string(nameBytes)
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
		return
	}

	domain, err := cmd.findDefaultDomain()

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if domain.Guid == "" {
		cmd.ui.Failed("No default domain exists")
	}

	return
}

func (cmd *Push) findDefaultDomain() (domain cf.Domain, err error) {
	var foundSharedDomain bool = false
	listDomainsCallback := api.ListDomainsCallback(func(domains []cf.Domain) bool {
		for _, aDomain := range domains {
			if aDomain.Shared {
				foundSharedDomain = true
				domain = aDomain
				break
			}
		}

		if foundSharedDomain {
			return false
		} else {
			return true
		}
	})

	apiResponse := cmd.domainRepo.ListSharedDomains(listDomainsCallback)
	if apiResponse.IsNotFound() {
		apiResponse = cmd.domainRepo.ListDomains(listDomainsCallback)
	}

	if apiResponse.IsNotSuccessful() {
		err = errors.New(apiResponse.Message)
	}

	if foundSharedDomain != true {
		err = errors.New("Could not find a default domain")
		return
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

func (cmd *Push) createOrUpdateApp(appParams cf.AppParams) (app cf.Application) {
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

	var didCreate bool = false
	if apiResponse.IsNotFound() {
		app, apiResponse = cmd.createApp(appParams)
		if apiResponse.IsNotSuccessful() {
			cmd.ui.Failed(apiResponse.Message)
			return
		}
		didCreate = true
	}

	if !didCreate {
		app = cmd.updateApp(app, appParams)
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

	if appParams.Has("env") {
		mergedEnvVars := generic.Merge(generic.NewMap(app.EnvironmentVars), generic.NewMap(appParams.Get("env")))
		appParams.Set("env", mergedEnvVars)
	}

	var apiResponse net.ApiResponse
	updatedApp, apiResponse = cmd.appRepo.Update(app.Guid, appParams)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	return
}

func (cmd *Push) findAndValidateAppsToPush(c *cli.Context) (appSet cf.AppSet) {
	baseManifestPath, manifestFilename := cmd.manifestPathFromContext(c)
	m := cmd.instantiateManifest(c, baseManifestPath, manifestFilename)

	appParams, err := cf.NewAppParamsFromContext(c)
	if err != nil {
		cmd.ui.Failed("Error: %s", err)
		return
	}

	if !appParams.Has("name") && len(m.Applications) > 1 && appParams.Count() > 0 {
		cmd.ui.Failed("%s", "Incorrect Usage. Command line flags (except -f) cannot be applied when pushing multiple apps from a manifest file.")
		return
	}

	appSet, err = cmd.createAppSetFromContextAndManifest(c, appParams, m)
	if err != nil {
		cmd.ui.Failed("Error: %s", err)
	}

	return
}

func (cmd *Push) manifestPathFromContext(c *cli.Context) (basePath, manifestFilename string) {
	var err error
	basePath, manifestFilename, err = cmd.manifestRepo.ManifestPath(c.String("f"))

	if err != nil {
		cmd.ui.Failed("%s", err)
		return
	}

	return
}

func (cmd *Push) instantiateManifest(c *cli.Context, manifestPath, manifestFilename string) (m *manifest.Manifest) {
	fullManifestPath := filepath.Join(manifestPath, manifestFilename)

	if c.Bool("no-manifest") {
		m = manifest.NewEmptyManifest()
		return
	}

	m, errs := cmd.manifestRepo.ReadManifest(fullManifestPath)

	if !errs.Empty() {
		if os.IsNotExist(errs[0]) && c.String("f") == "" {
			m = manifest.NewEmptyManifest()
			return
		} else {
			cmd.ui.Failed("Error reading manifest file:\n%s", errs)
			return
		}
	}

	// update paths in manifests to be relative to its directory
	for _, app := range m.Applications {
		if app.Has("path") {
			path := app.Get("path").(string)
			if filepath.IsAbs(path) {
				path = filepath.Clean(path)
			} else {
				path = filepath.Join(manifestPath, path)
			}
			app.Set("path", path)
		}
	}

	cmd.ui.Say("Using manifest file %s\n", terminal.EntityNameColor(fullManifestPath))
	return
}

func (cmd *Push) createAppSetFromContextAndManifest(c *cli.Context, contextParams cf.AppParams, m *manifest.Manifest) (appSet cf.AppSet, err error) {
	if len(m.Applications) > 1 {
		if contextParams.Has("name") {
			var app cf.AppParams
			app, err = findAppWithNameInManifest(contextParams.Get("name").(string), m)

			if err != nil {
				cmd.ui.Failed(fmt.Sprintf("Could not find app named '%s' in manifest", contextParams.Get("name").(string)))
				return
			}

			m.Applications = cf.AppSet{app}
		}
	}

	appSet = make(cf.AppSet, 0, len(m.Applications))
	if len(m.Applications) == 0 {
		if !contextParams.Has("name") || contextParams.Get("name") == "" {
			cmd.ui.FailWithUsage(c, "push")
			return
		}
		appSet = append(appSet, contextParams)
	} else {
		for _, manifestAppParams := range m.Applications {
			appFields := cf.NewAppParams(generic.Merge(manifestAppParams, contextParams))
			appSet = append(appSet, appFields)
		}
	}

	for _, appParams := range appSet {
		if !appParams.Has("name") {
			err = errors.New("app name is a required field")
		}
		if !appParams.Has("path") {
			cwd, _ := os.Getwd()
			appParams.Set("path", cwd)
		}
	}

	return
}

func findAppWithNameInManifest(name string, m *manifest.Manifest) (app cf.AppParams, err error) {
	for _, appParams := range m.Applications {
		if appParams.Has("name") && appParams.Get("name") == name {
			app = appParams
			return
		}
	}

	err = errors.New("Could not find named app in manifest")
	return
}
