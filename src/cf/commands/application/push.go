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
	ui            terminal.UI
	config        configuration.Reader
	manifestRepo  manifest.ManifestRepository
	appStarter    ApplicationStarter
	appStopper    ApplicationStopper
	serviceBinder service.ServiceBinder
	appRepo       api.ApplicationRepository
	domainRepo    api.DomainRepository
	routeRepo     api.RouteRepository
	serviceRepo   api.ServiceRepository
	stackRepo     api.StackRepository
	appBitsRepo   api.ApplicationBitsRepository
	wordGenerator words.WordGenerator
}

func NewPush(ui terminal.UI, config configuration.Reader, manifestRepo manifest.ManifestRepository,
	starter ApplicationStarter, stopper ApplicationStopper, binder service.ServiceBinder,
	appRepo api.ApplicationRepository, domainRepo api.DomainRepository, routeRepo api.RouteRepository,
	stackRepo api.StackRepository, serviceRepo api.ServiceRepository, appBitsRepo api.ApplicationBitsRepository,
	wordGenerator words.WordGenerator) *Push {
	return &Push{
		ui:            ui,
		config:        config,
		manifestRepo:  manifestRepo,
		appStarter:    starter,
		appStopper:    stopper,
		serviceBinder: binder,
		appRepo:       appRepo,
		domainRepo:    domainRepo,
		routeRepo:     routeRepo,
		serviceRepo:   serviceRepo,
		stackRepo:     stackRepo,
		appBitsRepo:   appBitsRepo,
		wordGenerator: wordGenerator,
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

		if appParams.ServicesToBind != nil {
			cmd.bindAppToServices(*appParams.ServicesToBind, app)
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
		err = cmd.serviceBinder.BindApplication(app, serviceInstance)

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

	switch apiErr.(type) {
	case nil:
		cmd.ui.Say("Using route %s", terminal.EntityNameColor(route.URL()))
	case *errors.ModelNotFoundError:
		cmd.ui.Say("Creating route %s...", terminal.EntityNameColor(domain.UrlForHost(hostname)))

		route, apiErr = cmd.routeRepo.Create(hostname, domain.Guid)
		if apiErr != nil {
			cmd.ui.Failed(apiErr.Error())
		}

		cmd.ui.Ok()
		cmd.ui.Say("")
	default:
		cmd.ui.Failed(apiErr.Error())
	}

	if !app.HasRoute(route) {
		cmd.ui.Say("Binding %s to %s...", terminal.EntityNameColor(domain.UrlForHost(hostname)), terminal.EntityNameColor(app.Name))

		apiErr = cmd.routeRepo.Bind(route.Guid, app.Guid)
		switch apiErr := apiErr.(type) {
		case nil:
			cmd.ui.Ok()
			cmd.ui.Say("")
			return
		case errors.HttpError:
			if apiErr.ErrorCode() == errors.INVALID_RELATION {
				cmd.ui.Failed("The route %s is already in use.\nTIP: Change the hostname with -n HOSTNAME or use --random-route to generate a new route and then push again.", route.URL())
			}
		}
		cmd.ui.Failed(apiErr.Error())
	}
}

func (cmd Push) hostnameForApp(appParams models.AppParams, c *cli.Context) string {
	if c.Bool("no-hostname") {
		return ""
	}

	if appParams.Host != nil {
		return *appParams.Host
	} else if appParams.UseRandomHostname {
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
		app, _ = cmd.appStopper.ApplicationStop(app)
	}

	cmd.ui.Say("")

	if c.Bool("no-start") {
		return
	}

	if params.HealthCheckTimeout != nil {
		cmd.appStarter.SetStartTimeoutSeconds(*params.HealthCheckTimeout)
	}

	cmd.appStarter.ApplicationStart(app)
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
	case *errors.HttpNotFoundError:
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
	case *errors.ModelNotFoundError:
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

func (cmd *Push) findAndValidateAppsToPush(c *cli.Context) []models.AppParams {
	appsFromManifest := cmd.getAppParamsFromManifest(c)
	appFromContext := cmd.getAppParamsFromContext(c)
	return cmd.createAppSetFromContextAndManifest(appFromContext, appsFromManifest)
}

func (cmd *Push) getAppParamsFromManifest(c *cli.Context) []models.AppParams {
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

	m, err := cmd.manifestRepo.ReadManifest(path)

	if err != nil {
		if m.Path == "" && c.String("f") == "" {
			return []models.AppParams{}
		} else {
			cmd.ui.Failed("Error reading manifest file:\n%s", err)
		}
	}

	apps, err := m.Applications()
	if err != nil {
		cmd.ui.Failed("Error reading manifest file:\n%s", err)
	}

	cmd.ui.Say("Using manifest file %s\n", terminal.EntityNameColor(m.Path))
	return apps
}

func (cmd *Push) createAppSetFromContextAndManifest(contextApp models.AppParams, manifestApps []models.AppParams) (apps []models.AppParams) {
	var err error

	switch len(manifestApps) {
	case 0:
		err = addApp(&apps, contextApp)
	case 1:
		manifestApps[0].Merge(&contextApp)
		err = addApp(&apps, manifestApps[0])
	default:
		selectedAppName := contextApp.Name
		contextApp.Name = nil

		if !contextApp.IsEmpty() {
			cmd.ui.Failed("%s", "Incorrect Usage. Command line flags (except -f) cannot be applied when pushing multiple apps from a manifest file.")
		}

		if selectedAppName != nil {
			var manifestApp models.AppParams
			manifestApp, err = findAppWithNameInManifest(*selectedAppName, manifestApps)
			if err == nil {
				addApp(&apps, manifestApp)
			}
		} else {
			for _, manifestApp := range manifestApps {
				addApp(&apps, manifestApp)
			}
		}
	}

	if err != nil {
		cmd.ui.Failed("Error: %s", err)
	}

	return
}

func addApp(apps *[]models.AppParams, app models.AppParams) (err error) {
	if app.Name == nil {
		err = errors.New("App name is a required field")
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

	err = errors.NewWithFmt("Could not find app named '%s' in manifest", name)
	return
}

func (cmd *Push) getAppParamsFromContext(c *cli.Context) (appParams models.AppParams) {
	if len(c.Args()) > 0 {
		appParams.Name = &c.Args()[0]
	}

	appParams.NoRoute = c.Bool("no-route")
	appParams.UseRandomHostname = c.Bool("random-route")

	if c.String("n") != "" {
		hostname := c.String("n")
		appParams.Host = &hostname
	}

	if c.String("b") != "" {
		buildpack := c.String("b")
		appParams.BuildpackUrl = &buildpack
	}

	if c.String("c") != "" {
		command := c.String("c")
		appParams.Command = &command
	}

	if c.String("c") == "null" {
		emptyStr := ""
		appParams.Command = &emptyStr
	}

	if c.String("d") != "" {
		domain := c.String("d")
		appParams.Domain = &domain
	}

	if c.String("i") != "" {
		instances, err := strconv.Atoi(c.String("i"))
		if err != nil {
			cmd.ui.Failed("Error: %s", errors.NewWithFmt("Invalid instances param: %s\n%s", c.String("i"), err))
		}
		appParams.InstanceCount = &instances
	}

	if c.String("k") != "" {
		diskQuota, err := formatters.ToMegabytes(c.String("k"))
		if err != nil {
			cmd.ui.Failed("Error: %s", errors.NewWithFmt("Invalid disk quota param: %s\n%s", c.String("k"), err))
			return
		}
		appParams.DiskQuota = &diskQuota
	}

	if c.String("m") != "" {
		memory, err := formatters.ToMegabytes(c.String("m"))
		if err != nil {
			cmd.ui.Failed("Error: %s", errors.NewWithFmt("Invalid memory param: %s\n%s", c.String("m"), err))
			return
		}
		appParams.Memory = &memory
	}

	if c.String("p") != "" {
		path := c.String("p")
		appParams.Path = &path
	}

	if c.String("s") != "" {
		stackName := c.String("s")
		appParams.StackName = &stackName
	}

	if c.String("t") != "" {
		timeout, err := strconv.Atoi(c.String("t"))
		if err != nil {
			cmd.ui.Failed("Error: %s", errors.NewWithFmt("Invalid timeout param: %s\n%s", c.String("t"), err))
		}

		appParams.HealthCheckTimeout = &timeout
	}

	return
}
