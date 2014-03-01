package application

import (
	"cf/api"
	"cf/commands/service"
	"cf/configuration"
	cferrors "cf/errors"
	"cf/formatters"
	"cf/manifest"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Push struct {
	ui             terminal.UI
	config         configuration.Reader
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
	globalServices []models.ServiceInstance
}

func NewPush(ui terminal.UI, config configuration.Reader, manifestRepo manifest.ManifestRepository,
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
		cmd.fetchStackGuid(&appParams)

		app := cmd.createOrUpdateApp(appParams)

		cmd.bindAppToRoute(app, appParams, c)

		cmd.ui.Say("Uploading %s...", terminal.EntityNameColor(app.Name))

		apiResponse := cmd.appBitsRepo.UploadApp(app.Guid, *appParams.Path, cmd.describeUploadOperation)
		if apiResponse != nil {
			cmd.ui.Failed(fmt.Sprintf("Error uploading application.\n%s", apiResponse.Error()))
			return
		}
		cmd.ui.Ok()

		if appParams.Services != nil {
			cmd.bindAppToServices(*appParams.Services, app)
		}

		cmd.restart(app, appParams, c)
	}
}

func (cmd *Push) bindAppToServices(services []string, app models.Application) {
	for _, serviceName := range services {
		serviceInstance, apiResponse := cmd.serviceRepo.FindInstanceByName(serviceName)

		if apiResponse != nil {
			cmd.ui.Failed("Could not find service %s to bind to %s", serviceName, app.Name)
			return
		}

		cmd.ui.Say("Binding service %s to %s in org %s / space %s as %s", serviceName, app.Name, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name, cmd.config.Username())
		apiResponse = cmd.binder.BindApplication(app, serviceInstance)

		if apiResponse != nil && apiResponse.ErrorCode() != service.AppAlreadyBoundErrorCode {
			cmd.ui.Failed("Could not find to service %s\nError: %s", serviceName, apiResponse)
			return
		}

		cmd.ui.Ok()
	}
}

func (cmd *Push) describeUploadOperation(path string, zipFileBytes, fileCount uint64) {
	humanReadableBytes := formatters.ByteSize(zipFileBytes)
	cmd.ui.Say("Uploading from: %s\n%s, %d files", path, humanReadableBytes, fileCount)
}

func (cmd *Push) fetchStackGuid(appParams *models.AppParams) {
	if appParams.StackName == nil {
		return
	}

	stackName := *appParams.StackName
	cmd.ui.Say("Using stack %s...", terminal.EntityNameColor(stackName))

	stack, apiResponse := cmd.stackRepo.FindByName(stackName)
	if apiResponse != nil {
		cmd.ui.Failed(apiResponse.Error())
		return
	}

	cmd.ui.Ok()
	appParams.StackGuid = &stack.Guid
}

func (cmd *Push) bindAppToRoute(app models.Application, params models.AppParams, c *cli.Context) {
	if c.Bool("no-route") {
		return
	}

	if params.NoRoute != nil && *params.NoRoute {
		cmd.ui.Say("App %s is a worker, skipping route creation", terminal.EntityNameColor(app.Name))
		return
	}

	routeFlagsPresent := c.String("n") != "" || c.String("d") != "" || c.Bool("no-hostname")
	if len(app.Routes) > 0 && !routeFlagsPresent {
		return
	}

	var defaultHostname string
	if params.Host != nil {
		defaultHostname = *params.Host
	} else {
		defaultHostname = hostNameForString(app.Name)
	}

	var domainName string
	if params.Domain != nil {
		domainName = *params.Domain
	} else {
		domainName = c.String("d")
	}

	hostName := cmd.hostname(c, defaultHostname)
	domain := cmd.domain(c, domainName)
	route := cmd.route(hostName, domain)

	for _, boundRoute := range app.Routes {
		if boundRoute.Guid == route.Guid {
			return
		}
	}

	cmd.ui.Say("Binding %s to %s...", terminal.EntityNameColor(domain.UrlForHost(hostName)), terminal.EntityNameColor(app.Name))

	apiResponse := cmd.routeRepo.Bind(route.Guid, app.Guid)
	if apiResponse != nil {
		cmd.ui.Failed(apiResponse.Error())
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

func (cmd *Push) restart(app models.Application, params models.AppParams, c *cli.Context) {
	if app.State != "stopped" {
		cmd.ui.Say("")
		app, _ = cmd.stopper.ApplicationStop(app)
	}

	cmd.ui.Say("")

	if c.Bool("no-start") {
		return
	}

	if params.HealthCheckTimeout != nil {
		cmd.starter.SetStartTimeoutSeconds(*params.HealthCheckTimeout)
	}

	cmd.starter.ApplicationStart(app)
}

func (cmd *Push) route(hostName string, domain models.DomainFields) (route models.Route) {
	route, apiResponse := cmd.routeRepo.FindByHostAndDomain(hostName, domain.Name)
	if apiResponse != nil {
		cmd.ui.Say("Creating route %s...", terminal.EntityNameColor(domain.UrlForHost(hostName)))

		route, apiResponse = cmd.routeRepo.Create(hostName, domain.Guid)
		if apiResponse != nil {
			cmd.ui.Failed(apiResponse.Error())
			return
		}

		cmd.ui.Ok()
		cmd.ui.Say("")
	} else {
		cmd.ui.Say("Using route %s", terminal.EntityNameColor(route.URL()))
	}

	return
}

func (cmd *Push) domain(c *cli.Context, domainName string) (domain models.DomainFields) {
	var apiResponse cferrors.Error

	if domainName != "" {
		domain, apiResponse = cmd.domainRepo.FindByNameInOrg(domainName, cmd.config.OrganizationFields().Guid)
		if apiResponse != nil {
			cmd.ui.Failed(apiResponse.Error())
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

func (cmd *Push) findDefaultDomain() (domain models.DomainFields, err error) {
	foundIt := false
	listDomainsCallback := func(aDomain models.DomainFields) bool {
		if aDomain.Shared {
			domain = aDomain
			foundIt = true
		}
		return !foundIt
	}

	apiResponse := cmd.domainRepo.ListSharedDomains(listDomainsCallback)
	if apiResponse != nil && apiResponse.IsNotFound() {
		apiResponse = cmd.domainRepo.ListDomains(listDomainsCallback)
	}

	if apiResponse != nil {
		err = errors.New(apiResponse.Error())
	}

	if !foundIt {
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

func (cmd *Push) createOrUpdateApp(appParams models.AppParams) (app models.Application) {
	if appParams.Name == nil {
		cmd.ui.Failed("Error: No name found for app")
		return
	}

	app, apiResponse := cmd.appRepo.Read(*appParams.Name)
	var didCreate bool = false

	if apiResponse != nil {
		if apiResponse.IsNotFound() {
			app, apiResponse = cmd.createApp(appParams)
			if apiResponse != nil {
				cmd.ui.Failed(apiResponse.Error())
				return
			}
			didCreate = true
		} else {
			cmd.ui.Failed(apiResponse.Error())
			return
		}
	}

	if !didCreate {
		app = cmd.updateApp(app, appParams)
	}

	return
}

func (cmd *Push) createApp(appParams models.AppParams) (app models.Application, apiResponse cferrors.Error) {
	spaceGuid := cmd.config.SpaceFields().Guid
	appParams.SpaceGuid = &spaceGuid

	cmd.ui.Say("Creating app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(*appParams.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	app, apiResponse = cmd.appRepo.Create(appParams)
	if apiResponse != nil {
		cmd.ui.Failed(apiResponse.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	return
}

func (cmd *Push) updateApp(app models.Application, appParams models.AppParams) (updatedApp models.Application) {
	cmd.ui.Say("Updating app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(app.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	if appParams.EnvironmentVars != nil {
		for key, val := range app.EnvironmentVars {
			if _, ok := (*appParams.EnvironmentVars)[key]; !ok {
				(*appParams.EnvironmentVars)[key] = val
			}
		}
	}

	var apiResponse cferrors.Error
	updatedApp, apiResponse = cmd.appRepo.Update(app.Guid, appParams)
	if apiResponse != nil {
		cmd.ui.Failed(apiResponse.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	return
}

func (cmd *Push) findAndValidateAppsToPush(c *cli.Context) (appSet []models.AppParams) {
	m := cmd.instantiateManifest(c)

	contextParams, err := newAppParamsFromContext(c)
	if err != nil {
		cmd.ui.Failed("Error: %s", err)
		return
	}

	if contextParams.Name == nil && len(m.Applications) > 1 && !contextParams.Equals(&models.AppParams{}) {
		cmd.ui.Failed("%s", "Incorrect Usage. Command line flags (except -f) cannot be applied when pushing multiple apps from a manifest file.")
		return
	}

	appSet, err = cmd.createAppSetFromContextAndManifest(c, contextParams, m)
	if err != nil {
		cmd.ui.Failed("Error: %s", err)
	}

	return
}

func (cmd *Push) instantiateManifest(c *cli.Context) (m *manifest.Manifest) {
	if c.Bool("no-manifest") {
		m = manifest.NewEmptyManifest()
		return
	}

	var path string
	if c.String("f") != "" {
		path = c.String("f")
	} else {
		var err error
		path, err = os.Getwd()
		if err != nil {
			cmd.ui.Failed("Could not determine the current working directory!", err)
			return
		}
	}

	m, manifestPath, errs := cmd.manifestRepo.ReadManifest(path)

	if !errs.Empty() {
		if manifestPath == "" && c.String("f") == "" {
			m = manifest.NewEmptyManifest()
		} else {
			cmd.ui.Failed("Error reading manifest file:\n%s", errs)
		}
		return
	}

	cmd.ui.Say("Using manifest file %s\n", terminal.EntityNameColor(manifestPath))
	return
}

func (cmd *Push) createAppSetFromContextAndManifest(c *cli.Context, contextParams models.AppParams, m *manifest.Manifest) (appSet []models.AppParams, err error) {
	if len(m.Applications) > 1 {
		if contextParams.Name != nil {
			var app models.AppParams
			app, err = findAppWithNameInManifest(*contextParams.Name, m)

			if err != nil {
				cmd.ui.Failed(fmt.Sprintf("Could not find app named '%s' in manifest", *contextParams.Name))
				return
			}

			m.Applications = []models.AppParams{app}
		}
	}

	appSet = make([]models.AppParams, 0, len(m.Applications))
	if len(m.Applications) == 0 {
		if contextParams.Name == nil || *contextParams.Name == "" {
			cmd.ui.FailWithUsage(c, "push")
			return
		}
		err = addApp(&appSet, contextParams)
	} else {
		for _, manifestAppParams := range m.Applications {
			manifestAppParams.Merge(&contextParams)
			err = addApp(&appSet, manifestAppParams)
		}
	}

	return
}

func addApp(apps *[]models.AppParams, app models.AppParams) (err error) {
	if app.Name == nil {
		err = errors.New("app name is a required field")
	}
	if app.Path == nil {
		cwd, _ := os.Getwd()
		app.Path = &cwd
	}
	*apps = append(*apps, app)
	return
}

func findAppWithNameInManifest(name string, m *manifest.Manifest) (app models.AppParams, err error) {
	for _, appParams := range m.Applications {
		if appParams.Name != nil && *appParams.Name == name {
			app = appParams
			return
		}
	}

	err = errors.New("Could not find named app in manifest")
	return
}

func newAppParamsFromContext(c *cli.Context) (appParams models.AppParams, err error) {
	if len(c.Args()) > 0 {
		appParams.Name = &c.Args()[0]
	}

	if c.String("b") != "" {
		buildpack := c.String("b")
		appParams.BuildpackUrl = &buildpack
	}

	if c.String("m") != "" {
		var memory uint64
		memory, err = formatters.ToMegabytes(c.String("m"))
		if err != nil {
			err = errors.New(fmt.Sprintf("Invalid memory param: %s\n%s", c.String("m"), err))
			return
		}
		appParams.Memory = &memory
	}

	if c.String("c") != "" {
		command := c.String("c")
		appParams.Command = &command
	}

	if c.String("c") == "null" {
		emptyStr := ""
		appParams.Command = &emptyStr
	}

	if c.String("i") != "" {
		var instances int
		instances, err = strconv.Atoi(c.String("i"))
		if err != nil {
			err = errors.New(fmt.Sprintf("Invalid instances param: %s\n%s", c.String("i"), err))
			return
		}
		appParams.InstanceCount = &instances
	}

	if c.String("s") != "" {
		stackName := c.String("s")
		appParams.StackName = &stackName
	}

	if c.String("t") != "" {
		var timeout int
		timeout, err = strconv.Atoi(c.String("t"))
		if err != nil {
			err = errors.New(fmt.Sprintf("Invalid timeout param: %s\n%s", c.String("t"), err))
			return
		}

		appParams.HealthCheckTimeout = &timeout
	}

	if c.String("p") != "" {
		var path string
		path, err = filepath.Abs(c.String("p"))
		if err != nil {
			err = errors.New(fmt.Sprintf("Error finding app path: %s", err))
			return
		}
		appParams.Path = &path
	}
	return
}
