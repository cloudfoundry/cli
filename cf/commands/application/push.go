package application

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/blang/semver"
	. "github.com/cloudfoundry/cli/cf/i18n"
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
	appfiles      app_files.AppFiles
}

func init() {
	command_registry.Register(&Push{})
}

func (cmd *Push) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["b"] = &cliFlags.StringFlag{ShortName: "b", Usage: T("Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'")}
	fs["c"] = &cliFlags.StringFlag{ShortName: "c", Usage: T("Startup command, set to null to reset to default start command")}
	fs["d"] = &cliFlags.StringFlag{ShortName: "d", Usage: T("Domain (e.g. example.com)")}
	fs["f"] = &cliFlags.StringFlag{ShortName: "f", Usage: T("Path to manifest")}
	fs["i"] = &cliFlags.IntFlag{ShortName: "i", Usage: T("Number of instances")}
	fs["k"] = &cliFlags.StringFlag{ShortName: "k", Usage: T("Disk limit (e.g. 256M, 1024M, 1G)")}
	fs["m"] = &cliFlags.StringFlag{ShortName: "m", Usage: T("Memory limit (e.g. 256M, 1024M, 1G)")}
	fs["hostname"] = &cliFlags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname (e.g. my-subdomain)")}
	fs["p"] = &cliFlags.StringFlag{ShortName: "p", Usage: T("Path to app directory or to a zip file of the contents of the app directory")}
	fs["s"] = &cliFlags.StringFlag{ShortName: "s", Usage: T("Stack to use (a stack is a pre-built file system, including an operating system, that can run apps)")}
	fs["t"] = &cliFlags.StringFlag{ShortName: "t", Usage: T("Maximum time (in seconds) for CLI to wait for application start, other server side timeouts may apply")}
	fs["docker-image"] = &cliFlags.StringFlag{Name: "docker-image", ShortName: "o", Usage: T("docker-image to be used (e.g. user/docker-image-name)")}
	fs["health-check-type"] = &cliFlags.StringFlag{Name: "health-check-type", ShortName: "u", Usage: T("Application health check type (e.g. port or none)")}
	fs["no-hostname"] = &cliFlags.BoolFlag{Name: "no-hostname", Usage: T("Map the root domain to this app")}
	fs["no-manifest"] = &cliFlags.BoolFlag{Name: "no-manifest", Usage: T("Ignore manifest file")}
	fs["no-route"] = &cliFlags.BoolFlag{Name: "no-route", Usage: T("Do not map a route to this app and remove routes from previous pushes of this app.")}
	fs["no-start"] = &cliFlags.BoolFlag{Name: "no-start", Usage: T("Do not start an app after pushing")}
	fs["random-route"] = &cliFlags.BoolFlag{Name: "random-route", Usage: T("Create a random route for this app")}
	fs["route-path"] = &cliFlags.StringFlag{Name: "route-path", Usage: T("Path for the route")}

	return command_registry.CommandMetadata{
		Name:        "push",
		ShortName:   "p",
		Description: T("Push a new app or sync changes to an existing app"),
		Usage: T("Push a single app (with or without a manifest):\n") + T("   CF_NAME push APP_NAME [-b BUILDPACK_NAME] [-c COMMAND] [-d DOMAIN] [-f MANIFEST_PATH] [--docker-image DOCKER_IMAGE]\n") + T("   [-i NUM_INSTANCES] [-k DISK] [-m MEMORY] [--hostname HOST] [-p PATH] [-s STACK] [-t TIMEOUT] [-u HEALTH_CHECK_TYPE] [--route-path ROUTE_PATH]\n") +
			"   [--no-hostname] [--no-manifest] [--no-route] [--no-start]\n" +
			"\n" + T("   Push multiple apps with a manifest:\n") + T("   CF_NAME push [-f MANIFEST_PATH]\n"),
		Flags: fs,
	}
}

func (cmd *Push) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) > 1 {
		cmd.ui.Failed(T("Incorrect Usage.\n\n") + command_registry.Commands.CommandUsage("push"))
	}

	var reqs []requirements.Requirement

	if fc.String("route-path") != "" {
		requiredVersion, err := semver.Make("2.36.0")
		if err != nil {
			panic(err.Error())
		}

		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--route-path'", requiredVersion))
	}

	reqs = append(reqs, []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}...)

	return reqs, nil
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
	cmd.appfiles = deps.AppFiles

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

		if c.IsSet("docker-image") {
			diego := true
			appParams.Diego = &diego
		}

		app := cmd.createOrUpdateApp(appParams)

		cmd.updateRoutes(routeActor, app, appParams)

		if c.String("docker-image") == "" {
			cmd.actor.ProcessPath(*appParams.Path, func(appDir string) {
				localFiles, err := cmd.appfiles.AppFilesInDir(appDir)
				if err != nil {
					cmd.ui.Failed(
						T("Error processing app files in '{{.Path}}': {{.Error}}",
							map[string]interface{}{
								"Path":  *appParams.Path,
								"Error": err.Error(),
							}),
					)
				}

				if len(localFiles) == 0 {
					cmd.ui.Failed(
						T("No app files found in '{{.Path}}'",
							map[string]interface{}{
								"Path": *appParams.Path,
							}),
					)
				}

				cmd.ui.Say(T("Uploading {{.AppName}}...",
					map[string]interface{}{"AppName": terminal.EntityNameColor(app.Name)}))

				apiErr := cmd.uploadApp(app.Guid, appDir, *appParams.Path, localFiles)
				if apiErr != nil {
					cmd.ui.Failed(fmt.Sprintf(T("Error uploading application.\n{{.ApiErr}}",
						map[string]interface{}{"ApiErr": apiErr.Error()})))
					return
				}
				cmd.ui.Ok()
			})
		}

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
		cmd.createAndBindRoute(nil, appParams.UseRandomHostname, routeActor, app, appParams.NoHostname, domain, appParams.RoutePath)
	} else {
		for _, host := range *(appParams.Hosts) {
			cmd.createAndBindRoute(&host, appParams.UseRandomHostname, routeActor, app, appParams.NoHostname, domain, appParams.RoutePath)
		}
	}
}

func (cmd *Push) createAndBindRoute(host *string, UseRandomHostname bool, routeActor actors.RouteActor, app models.Application, noHostName bool, domain models.DomainFields, routePath *string) {
	hostname := cmd.hostnameForApp(host, UseRandomHostname, app.Name, noHostName)
	var route models.Route
	if routePath != nil {
		route = routeActor.FindOrCreateRoute(hostname, domain, *routePath)
	} else {
		route = routeActor.FindOrCreateRoute(hostname, domain, "")
	}
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
	domain, err := cmd.domainRepo.FirstOrDefault(cmd.config.OrganizationFields().Guid, domainName)
	if err != nil {
		cmd.ui.Failed(err.Error())
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
		}
		cmd.ui.Failed(T("Error reading manifest file:\n{{.Err}}", map[string]interface{}{"Err": err.Error()}))
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
			cmd.ui.Failed(
				T("Manifest file is not found in the current directory, please provide either an app name or manifest") +
					"\n\n" +
					command_registry.Commands.CommandUsage("push"),
			)
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

	if c.String("route-path") != "" {
		routePath := c.String("route-path")
		appParams.RoutePath = &routePath
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

	if c.String("docker-image") != "" {
		dockerImage := c.String("docker-image")
		appParams.DockerImage = &dockerImage
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
			cmd.ui.Failed("Error: %s", fmt.Errorf(T("Invalid timeout param: {{.Timeout}}\n{{.Err}}",
				map[string]interface{}{"Timeout": c.String("t"), "Err": err.Error()})))
		}

		appParams.HealthCheckTimeout = &timeout
	}

	if healthCheckType := c.String("u"); healthCheckType != "" {
		if healthCheckType != "port" && healthCheckType != "none" {
			cmd.ui.Failed("Error: %s", fmt.Errorf(T("Invalid health-check-type param: {{.healthCheckType}}",
				map[string]interface{}{"healthCheckType": healthCheckType})))
		}

		appParams.HealthCheckType = &healthCheckType
	}

	return
}

func (cmd *Push) uploadApp(appGuid, appDir, appDirOrZipFile string, localFiles []models.AppFileFields) error {
	uploadDir, err := ioutil.TempDir("", "apps")
	defer os.RemoveAll(uploadDir)

	remoteFiles, hasFileToUpload, err := cmd.actor.GatherFiles(localFiles, appDir, uploadDir)
	if err != nil {
		return err
	}

	zipFile, err := ioutil.TempFile("", "uploads")
	defer func() {
		zipFile.Close()
		os.Remove(zipFile.Name())
	}()

	if hasFileToUpload {
		err = cmd.zipAppFiles(zipFile, appDirOrZipFile, uploadDir)
		if err != nil {
			return err
		}
	}

	err = cmd.actor.UploadApp(appGuid, zipFile, remoteFiles)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *Push) zipAppFiles(zipFile *os.File, appDir string, uploadDir string) error {
	err := cmd.zipper.Zip(uploadDir, zipFile)
	if err != nil {
		if emptyDirErr, ok := err.(*errors.EmptyDirError); ok {
			zipFile = nil
			return emptyDirErr
		}
		return fmt.Errorf("%s: %s", T("Error zipping application"), err.Error())
	}

	zipFileSize, err := cmd.zipper.GetZipSize(zipFile)
	if err != nil {
		return err
	}

	zipFileCount := cmd.appfiles.CountFiles(uploadDir)
	if zipFileCount > 0 {
		cmd.ui.Say(T("Uploading app files from: {{.Path}}", map[string]interface{}{"Path": appDir}))
		cmd.ui.Say(T("Uploading {{.ZipFileBytes}}, {{.FileCount}} files",
			map[string]interface{}{
				"ZipFileBytes": formatters.ByteSize(zipFileSize),
				"FileCount":    zipFileCount}))
	}

	return nil
}
