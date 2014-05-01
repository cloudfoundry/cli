package application

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/words"
	"github.com/codegangsta/cli"
	"os"
	"regexp"
	"strconv"
	"strings"
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
	authRepo      api.AuthenticationRepository
	wordGenerator words.WordGenerator
}

func NewPush(ui terminal.UI, config configuration.Reader, manifestRepo manifest.ManifestRepository,
	starter ApplicationStarter, stopper ApplicationStopper, binder service.ServiceBinder,
	appRepo api.ApplicationRepository, domainRepo api.DomainRepository, routeRepo api.RouteRepository,
	stackRepo api.StackRepository, serviceRepo api.ServiceRepository, appBitsRepo api.ApplicationBitsRepository,
	authRepo api.AuthenticationRepository, wordGenerator words.WordGenerator) *Push {
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
		authRepo:      authRepo,
		wordGenerator: wordGenerator,
	}
}

func (command *Push) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "push",
		ShortName:   "p",
		Description: "Push a new app or sync changes to an existing app",
		Usage: "Push a single app (with or without a manifest):\n" +
			"   CF_NAME push APP [-b BUILDPACK_NAME] [-c COMMAND] [-d DOMAIN] [-f MANIFEST_PATH]\n" +
			"   [-i NUM_INSTANCES] [-k DISK] [-m MEMORY] [-n HOST] [-p PATH] [-s STACK] [-t TIMEOUT]\n" +
			"   [--no-hostname] [--no-manifest] [--no-route] [--no-start]\n" +
			"\n" +
			"   Push multiple apps with a manifest:\n" +
			"   CF_NAME push [-f MANIFEST_PATH]\n",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("b", "Custom buildpack by name (e.g. my-buildpack) or GIT URL (e.g. https://github.com/heroku/heroku-buildpack-play.git)"),
			flag_helpers.NewStringFlag("c", "Startup command, set to null to reset to default start command"),
			flag_helpers.NewStringFlag("d", "Domain (e.g. example.com)"),
			flag_helpers.NewStringFlag("f", "Path to manifest"),
			flag_helpers.NewIntFlag("i", "Number of instances"),
			flag_helpers.NewStringFlag("k", "Disk limit (e.g. 256M, 1024M, 1G)"),
			flag_helpers.NewStringFlag("m", "Memory limit (e.g. 256M, 1024M, 1G)"),
			flag_helpers.NewStringFlag("n", "Hostname (e.g. my-subdomain)"),
			flag_helpers.NewStringFlag("p", "Path of app directory or zip file"),
			flag_helpers.NewStringFlag("s", "Stack to use (a stack is a pre-built file system, including an operating system, that can run apps)"),
			flag_helpers.NewStringFlag("t", "Start timeout in seconds"),
			cli.BoolFlag{Name: "no-hostname", Usage: "Map the root domain to this app"},
			cli.BoolFlag{Name: "no-manifest", Usage: "Ignore manifest file"},
			cli.BoolFlag{Name: "no-route", Usage: "Do not map a route to this app"},
			cli.BoolFlag{Name: "no-start", Usage: "Do not start an app after pushing"},
			cli.BoolFlag{Name: "random-route", Usage: "Create a random route for this app"},
		},
	}
}

func (cmd *Push) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) > 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "push")
		return
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd *Push) Run(c *cli.Context) {
	appSet := cmd.findAndValidateAppsToPush(c)
	cmd.authRepo.RefreshAuthToken()
	routeActor := actors.NewRouteActor(cmd.ui, cmd.routeRepo)
	noHostname := c.Bool("no-hostname")

	for _, appParams := range appSet {
		cmd.fetchStackGuid(&appParams)
		app := cmd.createOrUpdateApp(appParams)

		cmd.updateRoutes(routeActor, app, appParams, noHostname)

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

func (cmd *Push) updateRoutes(routeActor actors.RouteActor, app models.Application, appParams models.AppParams, noHostName bool) {
	defaultRouteAcceptable := len(app.Routes) == 0
	routeDefined := appParams.Domain != nil || appParams.Host != nil || noHostName

	domain := cmd.findDomain(appParams.Domain)
	hostname := cmd.hostnameForApp(appParams.Host, appParams.UseRandomHostname, app.Name, noHostName)

	if appParams.NoRoute {
		cmd.removeRoutes(app, routeActor)
	} else if routeDefined || defaultRouteAcceptable {
		route := routeActor.FindOrCreateRoute(hostname, domain)
		routeActor.BindRoute(app, route)
	}
}

func (cmd *Push) removeRoutes(app models.Application, routeActor actors.RouteActor) {
	if len(app.Routes) == 0 {
		cmd.ui.Say("App %s is a worker, skipping route creation", terminal.EntityNameColor(app.Name))
	} else {
		routeActor.UnbindAll(app)
	}
}

func (cmd *Push) hostnameForApp(host *string, useRandomHostName bool, name string, noHostName bool) string {
	if noHostName {
		return ""
	}

	if host != nil {
		return *host
	} else if useRandomHostName {
		return hostNameForString(name) + "-" + cmd.wordGenerator.Babble()
	} else {
		return hostNameForString(name)
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

func (cmd *Push) findDomain(domainName *string) (domain models.DomainFields) {
	domain, error := cmd.domainRepo.FirstOrDefault(cmd.config.OrganizationFields().Guid, domainName)
	if error != nil {
		cmd.ui.Failed(error.Error())
	}

	return
}

func (cmd *Push) bindAppToServices(services []string, app models.Application) {
	for _, serviceName := range services {
		serviceInstance, err := cmd.serviceRepo.FindInstanceByName(serviceName)

		if err != nil {
			cmd.ui.Failed("Could not find service %s to bind to %s", serviceName, app.Name)
			return
		}

		cmd.ui.Say("Binding service %s to app %s in org %s / space %s as %s...",
			terminal.EntityNameColor(serviceInstance.Name),
			terminal.EntityNameColor(app.Name),
			terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			terminal.EntityNameColor(cmd.config.Username()),
		)
		err = cmd.serviceBinder.BindApplication(app, serviceInstance)

		switch httpErr := err.(type) {
		case errors.HttpError:
			if httpErr.ErrorCode() == errors.APP_ALREADY_BOUND {
				err = nil
			}
		}

		if err != nil {
			cmd.ui.Failed("Could not bind to service %s\nError: %s", serviceName, err)
		}

		cmd.ui.Ok()
	}
}

func (cmd *Push) describeUploadOperation(path string, zipFileBytes, fileCount uint64) {
	if fileCount > 0 {
		cmd.ui.Say("Uploading app files from: %s", path)
		cmd.ui.Say("Uploading %s, %d files", formatters.ByteSize(zipFileBytes), fileCount)
	} else {
		cmd.ui.Warn("None of your application files have changed. Nothing will be uploaded.")
	}
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
		cmd.appStarter.SetStartTimeoutInSeconds(*params.HealthCheckTimeout)
	}

	cmd.appStarter.ApplicationStart(app)
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
		app = cmd.createApp(appParams)
	default:
		cmd.ui.Failed(apiErr.Error())
	}

	return
}

func (cmd *Push) createApp(appParams models.AppParams) (app models.Application) {
	spaceGuid := cmd.config.SpaceFields().Guid
	appParams.SpaceGuid = &spaceGuid

	cmd.ui.Say("Creating app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(*appParams.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	app, apiErr := cmd.appRepo.Create(appParams)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
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
		if buildpack == "null" || buildpack == "default" {
			buildpack = ""
		}
		appParams.BuildpackUrl = &buildpack
	}

	if c.String("c") != "" {
		command := c.String("c")
		if command == "null" || command == "default" {
			command = ""
		}
		appParams.Command = &command
	}

	if c.String("d") != "" {
		domain := c.String("d")
		appParams.Domain = &domain
	}

	if c.IsSet("i") {
		instances := c.Int("i")
		if instances < 1 {
			cmd.ui.Failed("Invalid instance count: %d\nInstance count must be a positive integer", instances)
		}
		appParams.InstanceCount = &instances
	}

	if c.String("k") != "" {
		diskQuota, err := formatters.ToMegabytes(c.String("k"))
		if err != nil {
			cmd.ui.Failed("Invalid disk quota: %s\n%s", c.String("k"), err)
		}
		appParams.DiskQuota = &diskQuota
	}

	if c.String("m") != "" {
		memory, err := formatters.ToMegabytes(c.String("m"))
		if err != nil {
			cmd.ui.Failed("Invalid memory limit: %s\n%s", c.String("m"), err)
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
