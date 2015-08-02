package application

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"

	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/api/stacks"
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/words/generator"
)

type Push struct {
	ui            terminal.UI
	config        core_config.Reader
	manifestRepo  manifest.ManifestRepository
	appStarter    ApplicationStarter
	appStopper    ApplicationStopper
	serviceBinder service.ServiceBinder
	appRepo       applications.ApplicationRepository
	domainRepo    api.DomainRepository
	routeRepo     api.RouteRepository
	serviceRepo   api.ServiceRepository
	stackRepo     stacks.StackRepository
	authRepo      authentication.AuthenticationRepository
	wordGenerator generator.WordGenerator
	actor         actors.PushActor
	zipper        app_files.Zipper
	app_files     app_files.AppFiles
}

func init() {
	command_registry.Register(&Push{})
}

func (cmd *Push) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["b"] = &cliFlags.StringFlag{Name: "b", Usage: T("Custom buildpack by name (e.g. my-buildpack) or GIT URL (e.g. 'https://github.com/heroku/heroku-buildpack-play.git') or GIT BRANCH URL (e.g. 'https://github.com/heroku/heroku-buildpack-play.git#develop' for 'develop' branch). Use built-in buildpacks only by setting value to 'null' or 'default'")}
	fs["c"] = &cliFlags.StringFlag{Name: "c", Usage: T("Startup command, set to null to reset to default start command")}
	fs["d"] = &cliFlags.StringFlag{Name: "d", Usage: T("Domain (e.g. example.com)")}
	fs["f"] = &cliFlags.StringFlag{Name: "f", Usage: T("Path to manifest")}
	fs["i"] = &cliFlags.IntFlag{Name: "i", Usage: T("Number of instances")}
	fs["k"] = &cliFlags.StringFlag{Name: "k", Usage: T("Disk limit (e.g. 256M, 1024M, 1G)")}
	fs["m"] = &cliFlags.StringFlag{Name: "m", Usage: T("Memory limit (e.g. 256M, 1024M, 1G)")}
	fs["n"] = &cliFlags.StringFlag{Name: "n", Usage: T("Hostname (e.g. my-subdomain)")}
	fs["p"] = &cliFlags.StringFlag{Name: "p", Usage: T("Path to app directory or to a zip file of the contents of the app directory")}
	fs["s"] = &cliFlags.StringFlag{Name: "s", Usage: T("Stack to use (a stack is a pre-built file system, including an operating system, that can run apps)")}
	fs["t"] = &cliFlags.StringFlag{Name: "t", Usage: T("Maximum time (in seconds) for CLI to wait for application start, other server side timeouts may apply")}
	fs["no-hostname"] = &cliFlags.BoolFlag{Name: "no-hostname", Usage: T("Map the root domain to this app")}
	fs["no-manifest"] = &cliFlags.BoolFlag{Name: "no-manifest", Usage: T("Ignore manifest file")}
	fs["no-route"] = &cliFlags.BoolFlag{Name: "no-route", Usage: T("Do not map a route to this app and remove routes from previous pushes of this app.")}
	fs["no-start"] = &cliFlags.BoolFlag{Name: "no-start", Usage: T("Do not start an app after pushing")}
	fs["random-route"] = &cliFlags.BoolFlag{Name: "random-route", Usage: T("Create a random route for this app")}

	return command_registry.CommandMetadata{
		Name:        "push",
		ShortName:   "p",
		Description: T("Push a new app or sync changes to an existing app"),
		Usage: T("Push a single app (with or without a manifest):\n") + T("   CF_NAME push APP_NAME [-b BUILDPACK_NAME] [-c COMMAND] [-d DOMAIN] [-f MANIFEST_PATH]\n") + T("   [-i NUM_INSTANCES] [-k DISK] [-m MEMORY] [-n HOST] [-p PATH] [-s STACK] [-t TIMEOUT]\n") +
			"   [--no-hostname] [--no-manifest] [--no-route] [--no-start]\n" +
			"\n" + T("   Push multiple apps with a manifest:\n") + T("   CF_NAME push [-f MANIFEST_PATH]\n"),
		Flags: fs,
	}
}

func (cmd *Push) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) > 1 {
		cmd.ui.Failed(T("Incorrect Usage.\n\n") + command_registry.Commands.CommandUsage("push"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}
	return
}

func (cmd *Push) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.manifestRepo = deps.ManifestRepo

	//set appStarter
	appCommand := command_registry.Commands.FindCommand("start")
	appCommand = appCommand.SetDependency(deps, false)
	cmd.appStarter = appCommand.(ApplicationStarter)

	//set appStopper
	appCommand = command_registry.Commands.FindCommand("stop")
	appCommand = appCommand.SetDependency(deps, false)
	cmd.appStopper = appCommand.(ApplicationStopper)

	//set serviceBinder
	appCommand = command_registry.Commands.FindCommand("bind-service")
	appCommand = appCommand.SetDependency(deps, false)
	cmd.serviceBinder = appCommand.(service.ServiceBinder)

	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	cmd.stackRepo = deps.RepoLocator.GetStackRepository()
	cmd.authRepo = deps.RepoLocator.GetAuthenticationRepository()
	cmd.wordGenerator = deps.WordGenerator
	cmd.actor = deps.PushActor
	cmd.zipper = deps.AppZipper
	cmd.app_files = deps.AppFiles

	return cmd
}

func (cmd *Push) Execute(c flags.FlagContext) {
	appSet := cmd.findAndValidateAppsToPush(c)
	_, apiErr := cmd.authRepo.RefreshAuthToken()
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	routeActor := actors.NewRouteActor(cmd.ui, cmd.routeRepo)

	for _, appParams := range appSet {
		cmd.fetchStackGuid(&appParams)
		app := cmd.createOrUpdateApp(appParams)

		cmd.updateRoutes(routeActor, app, appParams)

		cmd.ui.Say(T("Uploading {{.AppName}}...",
			map[string]interface{}{"AppName": terminal.EntityNameColor(app.Name)}))

		apiErr := cmd.uploadApp(app.Guid, *appParams.Path)
		if apiErr != nil {
			cmd.ui.Failed(fmt.Sprintf(T("Error uploading application.\n{{.ApiErr}}",
				map[string]interface{}{"ApiErr": apiErr.Error()})))
			return
		}
		cmd.ui.Ok()

		if appParams.ServicesToBind != nil {
			cmd.bindAppToServices(*appParams.ServicesToBind, app)
		}

		cmd.restart(app, appParams, c)
	}
}

func (cmd *Push) updateRoutes(routeActor actors.RouteActor, app models.Application, appParams models.AppParams) {
	defaultRouteAcceptable := len(app.Routes) == 0
	routeDefined := appParams.Domains != nil || !appParams.IsHostEmpty() || appParams.NoHostname

	if appParams.NoRoute {
		cmd.removeRoutes(app, routeActor)
		return
	}

	if routeDefined || defaultRouteAcceptable {
		if appParams.Domains == nil {
			cmd.processDomainsAndBindRoutes(appParams, routeActor, app, cmd.findDomain(nil))
		} else {
			for _, d := range *(appParams.Domains) {
				cmd.processDomainsAndBindRoutes(appParams, routeActor, app, cmd.findDomain(&d))
			}
		}
	}
}

func (cmd *Push) processDomainsAndBindRoutes(appParams models.AppParams, routeActor actors.RouteActor, app models.Application, domain models.DomainFields) {
	if appParams.IsHostEmpty() {
		cmd.createAndBindRoute(nil, appParams.UseRandomHostname, routeActor, app, appParams.NoHostname, domain)
	} else {
		for _, host := range *(appParams.Hosts) {
			cmd.createAndBindRoute(&host, appParams.UseRandomHostname, routeActor, app, appParams.NoHostname, domain)
		}
	}
}

func (cmd *Push) createAndBindRoute(host *string, UseRandomHostname bool, routeActor actors.RouteActor, app models.Application, noHostName bool, domain models.DomainFields) {
	hostname := cmd.hostnameForApp(host, UseRandomHostname, app.Name, noHostName)
	route := routeActor.FindOrCreateRoute(hostname, domain)
	routeActor.BindRoute(app, route)
}

func (cmd *Push) removeRoutes(app models.Application, routeActor actors.RouteActor) {
	if len(app.Routes) == 0 {
		cmd.ui.Say(T("App {{.AppName}} is a worker, skipping route creation",
			map[string]interface{}{"AppName": terminal.EntityNameColor(app.Name)}))
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
			cmd.ui.Failed(T("Could not find service {{.ServiceName}} to bind to {{.AppName}}",
				map[string]interface{}{"ServiceName": serviceName, "AppName": app.Name}))
			return
		}

		cmd.ui.Say(T("Binding service {{.ServiceName}} to app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
			map[string]interface{}{
				"ServiceName": terminal.EntityNameColor(serviceInstance.Name),
				"AppName":     terminal.EntityNameColor(app.Name),
				"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"Username":    terminal.EntityNameColor(cmd.config.Username())}))

		err = cmd.serviceBinder.BindApplication(app, serviceInstance, nil)

		switch httpErr := err.(type) {
		case errors.HttpError:
			if httpErr.ErrorCode() == errors.APP_ALREADY_BOUND {
				err = nil
			}
		}

		if err != nil {
			cmd.ui.Failed(T("Could not bind to service {{.ServiceName}}\nError: {{.Err}}",
				map[string]interface{}{"ServiceName": serviceName, "Err": err.Error()}))
		}

		cmd.ui.Ok()
	}
}

func (cmd *Push) fetchStackGuid(appParams *models.AppParams) {
	if appParams.StackName == nil {
		return
	}

	stackName := *appParams.StackName
	cmd.ui.Say(T("Using stack {{.StackName}}...",
		map[string]interface{}{"StackName": terminal.EntityNameColor(stackName)}))

	stack, apiErr := cmd.stackRepo.FindByName(stackName)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	appParams.StackGuid = &stack.Guid
}

func (cmd *Push) restart(app models.Application, params models.AppParams, c flags.FlagContext) {
	if app.State != T("stopped") {
		cmd.ui.Say("")
		app, _ = cmd.appStopper.ApplicationStop(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
	}

	cmd.ui.Say("")

	if c.Bool("no-start") {
		return
	}

	if params.HealthCheckTimeout != nil {
		cmd.appStarter.SetStartTimeoutInSeconds(*params.HealthCheckTimeout)
	}

	cmd.appStarter.ApplicationStart(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
}

func (cmd *Push) createOrUpdateApp(appParams models.AppParams) (app models.Application) {
	if appParams.Name == nil {
		cmd.ui.Failed(T("Error: No name found for app"))
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

	cmd.ui.Say(T("Creating app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(*appParams.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	app, apiErr := cmd.appRepo.Create(appParams)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	return
}

func (cmd *Push) updateApp(app models.Application, appParams models.AppParams) (updatedApp models.Application) {
	cmd.ui.Say(T("Updating app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

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

func (cmd *Push) findAndValidateAppsToPush(c flags.FlagContext) []models.AppParams {
	appsFromManifest := cmd.getAppParamsFromManifest(c)
	appFromContext := cmd.getAppParamsFromContext(c)
	return cmd.createAppSetFromContextAndManifest(appFromContext, appsFromManifest)
}

func (cmd *Push) getAppParamsFromManifest(c flags.FlagContext) []models.AppParams {
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
			cmd.ui.Failed(T("Could not determine the current working directory!"), err)
		}
	}

	m, err := cmd.manifestRepo.ReadManifest(path)

	if err != nil {
		if m.Path == "" && c.String("f") == "" {
			return []models.AppParams{}
		} else {
			cmd.ui.Failed(T("Error reading manifest file:\n{{.Err}}", map[string]interface{}{"Err": err.Error()}))
		}
	}

	apps, err := m.Applications()
	if err != nil {
		cmd.ui.Failed("Error reading manifest file:\n%s", err)
	}

	cmd.ui.Say(T("Using manifest file {{.Path}}\n",
		map[string]interface{}{"Path": terminal.EntityNameColor(m.Path)}))
	return apps
}

func (cmd *Push) createAppSetFromContextAndManifest(contextApp models.AppParams, manifestApps []models.AppParams) (apps []models.AppParams) {
	var err error

	switch len(manifestApps) {
	case 0:
		if contextApp.Name == nil {
			err = errors.New(T("Manifest file is not found in the current directory, please provide either an app name or manifest"))
		} else {
			err = addApp(&apps, contextApp)
		}
	case 1:
		manifestApps[0].Merge(&contextApp)
		err = addApp(&apps, manifestApps[0])
	default:
		selectedAppName := contextApp.Name
		contextApp.Name = nil

		if !contextApp.IsEmpty() {
			cmd.ui.Failed("%s", T("Incorrect Usage. Command line flags (except -f) cannot be applied when pushing multiple apps from a manifest file."))
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
		cmd.ui.Failed(T("Error: {{.Err}}", map[string]interface{}{"Err": err.Error()}))
	}

	return
}

func addApp(apps *[]models.AppParams, app models.AppParams) (err error) {
	if app.Name == nil {
		err = errors.New(T("App name is a required field"))
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

	err = errors.New(T("Could not find app named '{{.AppName}}' in manifest",
		map[string]interface{}{"AppName": name}))
	return
}

func (cmd *Push) getAppParamsFromContext(c flags.FlagContext) (appParams models.AppParams) {
	if len(c.Args()) > 0 {
		appParams.Name = &c.Args()[0]
	}

	appParams.NoRoute = c.Bool("no-route")
	appParams.UseRandomHostname = c.Bool("random-route")
	appParams.NoHostname = c.Bool("no-hostname")

	if c.String("n") != "" {
		appParams.Hosts = &[]string{c.String("n")}
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
		appParams.Domains = &[]string{c.String("d")}
	}

	if c.IsSet("i") {
		instances := c.Int("i")
		if instances < 1 {
			cmd.ui.Failed(T("Invalid instance count: {{.InstancesCount}}\nInstance count must be a positive integer",
				map[string]interface{}{"InstancesCount": instances}))
		}
		appParams.InstanceCount = &instances
	}

	if c.String("k") != "" {
		diskQuota, err := formatters.ToMegabytes(c.String("k"))
		if err != nil {
			cmd.ui.Failed(T("Invalid disk quota: {{.DiskQuota}}\n{{.Err}}",
				map[string]interface{}{"DiskQuota": c.String("k"), "Err": err.Error()}))
		}
		appParams.DiskQuota = &diskQuota
	}

	if c.String("m") != "" {
		memory, err := formatters.ToMegabytes(c.String("m"))
		if err != nil {
			cmd.ui.Failed(T("Invalid memory limit: {{.MemLimit}}\n{{.Err}}",
				map[string]interface{}{"MemLimit": c.String("m"), "Err": err.Error()}))
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
			cmd.ui.Failed("Error: %s", errors.NewWithFmt(T("Invalid timeout param: {{.Timeout}}\n{{.Err}}",
				map[string]interface{}{"Timeout": c.String("t"), "Err": err.Error()})))
		}

		appParams.HealthCheckTimeout = &timeout
	}

	return
}

func (cmd *Push) uploadApp(appGuid string, appDir string) (apiErr error) {
	fileutils.TempDir("apps", func(uploadDir string, err error) {
		if err != nil {
			apiErr = err
			return
		}

		presentFiles, hasFileToUpload, err := cmd.actor.GatherFiles(appDir, uploadDir)
		if err != nil {
			apiErr = err
			return
		}

		fileutils.TempFile("uploads", func(zipFile *os.File, err error) {
			if hasFileToUpload {
				err = cmd.zipAppFiles(zipFile, appDir, uploadDir)
				if err != nil {
					apiErr = err
					return
				}
			}

			err = cmd.actor.UploadApp(appGuid, zipFile, presentFiles)
			if err != nil {
				apiErr = err
				return
			}
		})
		return
	})
	return
}

func (cmd *Push) zipAppFiles(zipFile *os.File, appDir string, uploadDir string) (zipErr error) {
	zipErr = cmd.zipWithBetterErrors(uploadDir, zipFile)
	if zipErr != nil {
		return
	}

	zipFileSize, zipErr := cmd.zipper.GetZipSize(zipFile)
	if zipErr != nil {
		return
	}

	zipFileCount := cmd.app_files.CountFiles(uploadDir)

	cmd.describeUploadOperation(appDir, zipFileSize, zipFileCount)
	return
}

func (cmd *Push) zipWithBetterErrors(uploadDir string, zipFile *os.File) error {
	zipError := cmd.zipper.Zip(uploadDir, zipFile)
	switch err := zipError.(type) {
	case nil:
		return nil
	case *errors.EmptyDirError:
		zipFile = nil
		return zipError
	default:
		return errors.NewWithError(T("Error zipping application"), err)
	}
}

func (cmd *Push) describeUploadOperation(path string, zipFileBytes, fileCount int64) {
	if fileCount > 0 {
		cmd.ui.Say(T("Uploading app files from: {{.Path}}", map[string]interface{}{"Path": path}))
		cmd.ui.Say(T("Uploading {{.ZipFileBytes}}, {{.FileCount}} files",
			map[string]interface{}{
				"ZipFileBytes": formatters.ByteSize(zipFileBytes),
				"FileCount":    fileCount}))
	}
}
