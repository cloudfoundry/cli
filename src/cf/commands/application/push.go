package application

import (
	"cf/api"
	"cf/commands/service"
	"cf/configuration"
	"cf/errors"
	"cf/formatters"
	"cf/manifest"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"regexp"
	"strconv"
	"strings"
	"words"
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
	wordGenerator  words.WordGenerator
}

func NewPush(ui terminal.UI, config configuration.Reader, manifestRepo manifest.ManifestRepository,
	starter ApplicationStarter, stopper ApplicationStopper, binder service.ServiceBinder,
	appRepo api.ApplicationRepository, domainRepo api.DomainRepository, routeRepo api.RouteRepository,
	stackRepo api.StackRepository, serviceRepo api.ServiceRepository, appBitsRepo api.ApplicationBitsRepository,
	wordGenerator words.WordGenerator) (cmd *Push) {
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
	cmd.wordGenerator = wordGenerator
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

		apiErr := cmd.appBitsRepo.UploadApp(app.Guid, *appParams.Path, cmd.describeUploadOperation)
		if apiErr != nil {
			cmd.ui.Failed(fmt.Sprintf("Error uploading application.\n%s", apiErr.Error()))
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
		serviceInstance, err := cmd.serviceRepo.FindInstanceByName(serviceName)

		if err != nil {
			cmd.ui.Failed("Could not find service %s to bind to %s", serviceName, app.Name)
			return
		}

		cmd.ui.Say("Binding service %s to %s in org %s / space %s as %s", serviceName, app.Name, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name, cmd.config.Username())
		err = cmd.binder.BindApplication(app, serviceInstance)

		if err, ok := err.(errors.HttpError); ok && err.ErrorCode() == service.AppAlreadyBoundErrorCode {
			err = nil
		}

		if err != nil {
			cmd.ui.Failed("Could not find to service %s\nError: %s", serviceName, err)
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

	stack, apiErr := cmd.stackRepo.FindByName(stackName)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	appParams.StackGuid = &stack.Guid
}

func (cmd *Push) bindAppToRoute(app models.Application, params models.AppParams, c *cli.Context) {
	if params.NoRoute {
		cmd.ui.Say("App %s is a worker, skipping route creation", terminal.EntityNameColor(app.Name))
		return
	}

	routeFlagsPresent := c.String("n") != "" || c.String("d") != "" || c.Bool("no-hostname")
	if len(app.Routes) > 0 && !routeFlagsPresent {
		return
	}

	domain := cmd.findDomain(params)
	hostname := cmd.hostnameForApp(params, c)

	route, apiErr := cmd.routeRepo.FindByHostAndDomain(hostname, domain.Name)

	// FIXME: ensure this is a NotFound error
	if apiErr != nil {
		cmd.ui.Say("Creating route %s...", terminal.EntityNameColor(domain.UrlForHost(hostname)))

		route, apiErr = cmd.routeRepo.Create(hostname, domain.Guid)
		if apiErr != nil {
			cmd.ui.Failed(apiErr.Error())
		}

		cmd.ui.Ok()
		cmd.ui.Say("")
	} else {
		cmd.ui.Say("Using route %s", terminal.EntityNameColor(route.URL()))
	}

	if !app.HasRoute(route) {
		cmd.ui.Say("Binding %s to %s...", terminal.EntityNameColor(domain.UrlForHost(hostname)), terminal.EntityNameColor(app.Name))

		apiErr = cmd.routeRepo.Bind(route.Guid, app.Guid)
		if apiErr != nil {
			cmd.ui.Failed(apiErr.Error())
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say("")
}

func (cmd Push) hostnameForApp(appParams models.AppParams, c *cli.Context) string {
	if c.Bool("no-hostname") {
		return ""
	}

	if appParams.Host != nil {
		return *appParams.Host
	} else if appParams.RandomHostname {
		return hostNameForString(*appParams.Name) + "-" + cmd.wordGenerator.Babble()
	} else {
		return hostNameForString(*appParams.Name)
	}
}

var forbiddenHostCharRegex = regexp.MustCompile("[^a-z0-9-]")
var whitespaceRegex = regexp.MustCompile(`[\s_]+`)

func hostNameForString(name string) string {
	name = strings.ToLower(name)
	name = whitespaceRegex.ReplaceAllString(name, "-")
	name = forbiddenHostCharRegex.ReplaceAllString(name, "")
	return name
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

func (cmd *Push) findDomain(appParams models.AppParams) (domain models.DomainFields) {
	var err error
	if appParams.Domain != nil {
		domain, err = cmd.domainRepo.FindByNameInOrg(*appParams.Domain, cmd.config.OrganizationFields().Guid)
		if err != nil {
			cmd.ui.Failed(err.Error())
		}
	} else {
		domain, err = cmd.findDefaultDomain()
		if err != nil {
			cmd.ui.Failed(err.Error())
		}
		if domain.Guid == "" {
			cmd.ui.Failed("No default domain exists")
		}
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

	apiErr := cmd.domainRepo.ListSharedDomains(listDomainsCallback)

	// FIXME: needs semantic API version
	switch apiErr.(type) {
	case errors.HttpNotFoundError:
		apiErr = cmd.domainRepo.ListDomains(listDomainsCallback)
	}

	if !foundIt {
		err = errors.New("Could not find a default domain")
		return
	}

	return
}

func (cmd *Push) createOrUpdateApp(appParams models.AppParams) (app models.Application) {
	if appParams.Name == nil {
		cmd.ui.Failed("Error: No name found for app")
	}

	app, apiErr := cmd.appRepo.Read(*appParams.Name)

	switch apiErr.(type) {
	case nil:
		app = cmd.updateApp(app, appParams)
	case errors.ModelNotFoundError:
		app, apiErr = cmd.createApp(appParams)
		if apiErr != nil {
			cmd.ui.Failed(apiErr.Error())
			return
		}
	default:
		cmd.ui.Failed(apiErr.Error())
		return
	}

	return
}

func (cmd *Push) createApp(appParams models.AppParams) (app models.Application, apiErr error) {
	spaceGuid := cmd.config.SpaceFields().Guid
	appParams.SpaceGuid = &spaceGuid

	cmd.ui.Say("Creating app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(*appParams.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	app, apiErr = cmd.appRepo.Create(appParams)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
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

	var apiErr error
	updatedApp, apiErr = cmd.appRepo.Update(app.Guid, appParams)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	return
}

func (cmd *Push) findAndValidateAppsToPush(c *cli.Context) (appSet []models.AppParams) {
	apps := cmd.instantiateManifest(c)

	contextParams, err := newAppParamsFromContext(c)
	if err != nil {
		cmd.ui.Failed("Error: %s", err)
		return
	}

	if contextParams.Name == nil && len(apps) > 1 && !contextParams.Equals(&models.AppParams{}) {
		cmd.ui.Failed("%s", "Incorrect Usage. Command line flags (except -f) cannot be applied when pushing multiple apps from a manifest file.")
		return
	}

	appSet, err = cmd.createAppSetFromContextAndManifest(c, contextParams, apps)
	if err != nil {
		cmd.ui.Failed("Error: %s", err)
	}

	return
}

func (cmd *Push) instantiateManifest(c *cli.Context) []models.AppParams {
	if c.Bool("no-manifest") {
		return []models.AppParams{}
	}

	var path string
	if c.String("f") != "" {
		path = c.String("f")
	} else {
		var err error
		path, err = os.Getwd()
		if err != nil {
			cmd.ui.Failed("Could not determine the current working directory!", err)
		}
	}

	m, errs := cmd.manifestRepo.ReadManifest(path)

	if !errs.Empty() {
		if m.Path == "" && c.String("f") == "" {
			return []models.AppParams{}
		} else {
			cmd.ui.Failed("Error reading manifest file:\n%s", errs)
		}
	}

	apps, errs := m.Applications()
	if !errs.Empty() {
		if m.Path == "" && c.String("f") == "" {
			return []models.AppParams{}
		} else {
			cmd.ui.Failed("Error reading manifest file:\n%s", errs)
		}
	}

	cmd.ui.Say("Using manifest file %s\n", terminal.EntityNameColor(m.Path))
	return apps
}

func (cmd *Push) createAppSetFromContextAndManifest(c *cli.Context, contextParams models.AppParams, manifestApps []models.AppParams) (appSet []models.AppParams, err error) {
	if len(manifestApps) > 1 {
		if contextParams.Name != nil {
			var app models.AppParams
			app, err = findAppWithNameInManifest(*contextParams.Name, manifestApps)

			if err != nil {
				cmd.ui.Failed(fmt.Sprintf("Could not find app named '%s' in manifest", *contextParams.Name))
				return
			}

			manifestApps = []models.AppParams{app}
		}
	}

	appSet = make([]models.AppParams, 0, len(manifestApps))
	if len(manifestApps) == 0 {
		if contextParams.Name == nil || *contextParams.Name == "" {
			cmd.ui.FailWithUsage(c, "push")
			return
		}
		err = addApp(&appSet, contextParams)
	} else {
		for _, manifestAppParams := range manifestApps {
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

func findAppWithNameInManifest(name string, manifestApps []models.AppParams) (app models.AppParams, err error) {
	for _, appParams := range manifestApps {
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

	appParams.NoRoute = c.Bool("no-route")
	appParams.RandomHostname = c.Bool("random-route")

	if c.String("n") != "" {
		hostname := c.String("n")
		appParams.Host = &hostname
	}

	if c.String("d") != "" {
		domain := c.String("d")
		appParams.Domain = &domain
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
		path := c.String("p")
		appParams.Path = &path
	}
	return
}
