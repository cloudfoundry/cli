package application

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/actors"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/api/authentication"
	"code.cloudfoundry.org/cli/cf/api/stacks"
	"code.cloudfoundry.org/cli/cf/appfiles"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/service"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/formatters"
	"code.cloudfoundry.org/cli/cf/manifest"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type Push struct {
	ui            terminal.UI
	config        coreconfig.Reader
	manifestRepo  manifest.Repository
	appStarter    Starter
	appStopper    Stopper
	serviceBinder service.Binder
	appRepo       applications.Repository
	domainRepo    api.DomainRepository
	routeRepo     api.RouteRepository
	serviceRepo   api.ServiceRepository
	stackRepo     stacks.StackRepository
	authRepo      authentication.Repository
	wordGenerator commandregistry.RandomWordGenerator
	actor         actors.PushActor
	routeActor    actors.RouteActor
	zipper        appfiles.Zipper
	appfiles      appfiles.AppFiles
}

func init() {
	commandregistry.Register(&Push{})
}

func (cmd *Push) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["b"] = &flags.StringFlag{ShortName: "b", Usage: T("Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'")}
	fs["c"] = &flags.StringFlag{ShortName: "c", Usage: T("Startup command, set to null to reset to default start command")}
	fs["d"] = &flags.StringFlag{ShortName: "d", Usage: T("Domain (e.g. example.com)")}
	fs["f"] = &flags.StringFlag{ShortName: "f", Usage: T("Path to manifest")}
	fs["i"] = &flags.IntFlag{ShortName: "i", Usage: T("Number of instances")}
	fs["k"] = &flags.StringFlag{ShortName: "k", Usage: T("Disk limit (e.g. 256M, 1024M, 1G)")}
	fs["m"] = &flags.StringFlag{ShortName: "m", Usage: T("Memory limit (e.g. 256M, 1024M, 1G)")}
	fs["hostname"] = &flags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname (e.g. my-subdomain)")}
	fs["p"] = &flags.StringFlag{ShortName: "p", Usage: T("Path to app directory or to a zip file of the contents of the app directory")}
	fs["s"] = &flags.StringFlag{ShortName: "s", Usage: T("Stack to use (a stack is a pre-built file system, including an operating system, that can run apps)")}
	fs["t"] = &flags.StringFlag{ShortName: "t", Usage: T("Time (in seconds) allowed to elapse between starting up an app and the first healthy response from the app")}
	fs["docker-image"] = &flags.StringFlag{Name: "docker-image", ShortName: "o", Usage: T("Docker-image to be used (e.g. user/docker-image-name)")}
	fs["docker-username"] = &flags.StringFlag{Name: "docker-username", Usage: T("Repository username; used with password from environment variable CF_DOCKER_PASSWORD")}
	fs["health-check-type"] = &flags.StringFlag{Name: "health-check-type", ShortName: "u", Usage: T("Application health check type (Default: 'port', 'none' accepted for 'process', 'http' implies endpoint '/')")}
	fs["no-hostname"] = &flags.BoolFlag{Name: "no-hostname", Usage: T("Map the root domain to this app")}
	fs["no-manifest"] = &flags.BoolFlag{Name: "no-manifest", Usage: T("Ignore manifest file")}
	fs["no-route"] = &flags.BoolFlag{Name: "no-route", Usage: T("Do not map a route to this app and remove routes from previous pushes of this app")}
	fs["no-start"] = &flags.BoolFlag{Name: "no-start", Usage: T("Do not start an app after pushing")}
	fs["random-route"] = &flags.BoolFlag{Name: "random-route", Usage: T("Create a random route for this app")}
	fs["route-path"] = &flags.StringFlag{Name: "route-path", Usage: T("Path for the route")}
	// Hidden:true to hide app-ports for release #117189491
	fs["app-ports"] = &flags.StringFlag{Name: "app-ports", Usage: T("Comma delimited list of ports the application may listen on"), Hidden: true}

	return commandregistry.CommandMetadata{
		Name:        "push",
		ShortName:   "p",
		Description: T("Push a new app or sync changes to an existing app"),
		// strings.Replace \\n with newline so this string matches the new usage string but still gets displayed correctly
		Usage: []string{strings.Replace(T("cf push APP_NAME [-b BUILDPACK_NAME] [-c COMMAND] [-f MANIFEST_PATH | --no-manifest] [--no-start]\\n   [-i NUM_INSTANCES] [-k DISK] [-m MEMORY] [-p PATH] [-s STACK] [-t HEALTH_TIMEOUT] [-u (process | port | http)]\\n   [--no-route | --random-route | --hostname HOST | --no-hostname] [-d DOMAIN] [--route-path ROUTE_PATH]\\n\\n   cf push APP_NAME --docker-image [REGISTRY_HOST:PORT/]IMAGE[:TAG] [--docker-username USERNAME]\\n   [-c COMMAND] [-f MANIFEST_PATH | --no-manifest] [--no-start]\\n   [-i NUM_INSTANCES] [-k DISK] [-m MEMORY] [-t HEALTH_TIMEOUT] [-u (process | port | http)]\\n   [--no-route | --random-route | --hostname HOST | --no-hostname] [-d DOMAIN] [--route-path ROUTE_PATH]\\n\\n   cf push -f MANIFEST_WITH_MULTIPLE_APPS_PATH [APP_NAME] [--no-start]"), "\\n", "\n", -1)},
		Flags: fs,
	}
}

func (cmd *Push) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	var reqs []requirements.Requirement

	usageReq := requirementsFactory.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd), "",
		func() bool {
			return len(fc.Args()) > 1
		},
	)

	reqs = append(reqs, usageReq)

	if fc.String("route-path") != "" {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--route-path'", cf.RoutePathMinimumAPIVersion))
	}

	if fc.String("app-ports") != "" {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--app-ports'", cf.MultipleAppPortsMinimumAPIVersion))
	}

	reqs = append(reqs, []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}...)

	return reqs, nil
}

func (cmd *Push) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.manifestRepo = deps.ManifestRepo

	//set appStarter
	appCommand := commandregistry.Commands.FindCommand("start")
	appCommand = appCommand.SetDependency(deps, false)
	cmd.appStarter = appCommand.(Starter)

	//set appStopper
	appCommand = commandregistry.Commands.FindCommand("stop")
	appCommand = appCommand.SetDependency(deps, false)
	cmd.appStopper = appCommand.(Stopper)

	//set serviceBinder
	appCommand = commandregistry.Commands.FindCommand("bind-service")
	appCommand = appCommand.SetDependency(deps, false)
	cmd.serviceBinder = appCommand.(service.Binder)

	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	cmd.stackRepo = deps.RepoLocator.GetStackRepository()
	cmd.authRepo = deps.RepoLocator.GetAuthenticationRepository()
	cmd.wordGenerator = deps.WordGenerator
	cmd.actor = deps.PushActor
	cmd.routeActor = deps.RouteActor
	cmd.zipper = deps.AppZipper
	cmd.appfiles = deps.AppFiles

	return cmd
}

func (cmd *Push) Execute(c flags.FlagContext) error {
	appsFromManifest, err := cmd.getAppParamsFromManifest(c)
	if err != nil {
		return err
	}

	errs := cmd.actor.ValidateAppParams(appsFromManifest)
	if len(errs) > 0 {
		errStr := T("Invalid application configuration") + ":"

		for _, e := range errs {
			errStr = fmt.Sprintf("%s\n%s", errStr, e.Error())
		}

		return fmt.Errorf("%s", errStr)
	}

	appFromContext, err := cmd.getAppParamsFromContext(c)
	if err != nil {
		return err
	}

	err = cmd.ValidateContextAndAppParams(appsFromManifest, appFromContext)
	if err != nil {
		return err
	}

	appSet, err := cmd.createAppSetFromContextAndManifest(appFromContext, appsFromManifest)
	if err != nil {
		return err
	}

	_, err = cmd.authRepo.RefreshAuthToken()
	if err != nil {
		return err
	}

	for _, appParams := range appSet {
		if appParams.Name == nil {
			return errors.New(T("Error: No name found for app"))
		}

		err = cmd.fetchStackGUID(&appParams)
		if err != nil {
			return err
		}

		if appParams.DockerImage != nil {
			diego := true
			appParams.Diego = &diego
		}

		var app, existingApp models.Application
		existingApp, err = cmd.appRepo.Read(*appParams.Name)
		switch err.(type) {
		case nil:
			cmd.ui.Say(T("Updating app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
				map[string]interface{}{
					"AppName":   terminal.EntityNameColor(existingApp.Name),
					"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
					"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
					"Username":  terminal.EntityNameColor(cmd.config.Username())}))

			if appParams.EnvironmentVars != nil {
				for key, val := range existingApp.EnvironmentVars {
					if _, ok := (*appParams.EnvironmentVars)[key]; !ok {
						(*appParams.EnvironmentVars)[key] = val
					}
				}
			}

			// if the user did not provide a health-check-http-endpoint
			// and one doesn't exist already in the cloud
			// set to default
			if appParams.HealthCheckType != nil && *appParams.HealthCheckType == "http" {
				if appParams.HealthCheckHTTPEndpoint == nil && existingApp.HealthCheckHTTPEndpoint == "" {
					endpoint := "/"
					appParams.HealthCheckHTTPEndpoint = &endpoint
				}
			}

			app, err = cmd.appRepo.Update(existingApp.GUID, appParams)
			if err != nil {
				return err
			}
		case *errors.ModelNotFoundError:
			spaceGUID := cmd.config.SpaceFields().GUID
			appParams.SpaceGUID = &spaceGUID

			cmd.ui.Say(T("Creating app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
				map[string]interface{}{
					"AppName":   terminal.EntityNameColor(*appParams.Name),
					"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
					"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
					"Username":  terminal.EntityNameColor(cmd.config.Username())}))

			// if the user did not provide a health-check-http-endpoint
			// set to default
			if appParams.HealthCheckType != nil && *appParams.HealthCheckType == "http" {
				if appParams.HealthCheckHTTPEndpoint == nil {
					endpoint := "/"
					appParams.HealthCheckHTTPEndpoint = &endpoint
				}
			}
			app, err = cmd.appRepo.Create(appParams)
			if err != nil {
				return err
			}
		default:
			return err
		}

		cmd.ui.Ok()
		cmd.ui.Say("")

		err = cmd.updateRoutes(app, appParams, appFromContext)
		if err != nil {
			return err
		}

		if appParams.DockerImage == nil {
			err = cmd.actor.ProcessPath(*appParams.Path, cmd.processPathCallback(*appParams.Path, app))
			if err != nil {
				return errors.New(
					T("Error processing app files: {{.Error}}",
						map[string]interface{}{
							"Error": err.Error(),
						}),
				)
			}
		}

		if appParams.ServicesToBind != nil {
			err = cmd.bindAppToServices(appParams.ServicesToBind, app)
			if err != nil {
				return err
			}
		}

		err = cmd.restart(app, appParams, c)
		if err != nil {
			return errors.New(
				T("Error restarting application: {{.Error}}",
					map[string]interface{}{
						"Error": err.Error(),
					}),
			)
		}
	}
	return nil
}

func (cmd *Push) processPathCallback(path string, app models.Application) func(string) error {
	return func(appDir string) error {
		localFiles, err := cmd.appfiles.AppFilesInDir(appDir)
		if err != nil {
			return errors.New(
				T("Error processing app files in '{{.Path}}': {{.Error}}",
					map[string]interface{}{
						"Path":  path,
						"Error": err.Error(),
					}))
		}

		if len(localFiles) == 0 {
			return errors.New(
				T("No app files found in '{{.Path}}'",
					map[string]interface{}{
						"Path": path,
					}))
		}

		cmd.ui.Say(T("Uploading {{.AppName}}...",
			map[string]interface{}{"AppName": terminal.EntityNameColor(app.Name)}))

		err = cmd.uploadApp(app.GUID, appDir, path, localFiles)
		if err != nil {
			return errors.New(T("Error uploading application.\n{{.APIErr}}",
				map[string]interface{}{"APIErr": err.Error()}))
		}
		cmd.ui.Ok()
		return nil
	}
}

func (cmd *Push) updateRoutes(app models.Application, appParams models.AppParams, appParamsFromContext models.AppParams) error {
	defaultRouteAcceptable := len(app.Routes) == 0
	routeDefined := appParams.Domains != nil || !appParams.IsHostEmpty() || appParams.IsNoHostnameTrue()

	switch {
	case appParams.NoRoute:
		if len(app.Routes) == 0 {
			cmd.ui.Say(T("App {{.AppName}} is a worker, skipping route creation",
				map[string]interface{}{"AppName": terminal.EntityNameColor(app.Name)}))
		} else {
			err := cmd.routeActor.UnbindAll(app)
			if err != nil {
				return err
			}
		}
	case len(appParams.Routes) > 0:
		for _, manifestRoute := range appParams.Routes {
			err := cmd.actor.MapManifestRoute(manifestRoute.Route, app, appParamsFromContext)
			if err != nil {
				return err
			}
		}
	case (routeDefined || defaultRouteAcceptable) && appParams.Domains == nil:
		domain, err := cmd.findDomain(nil)
		if err != nil {
			return err
		}
		appParams.UseRandomPort = isTCP(domain)
		err = cmd.processDomainsAndBindRoutes(appParams, app, domain)
		if err != nil {
			return err
		}
	case routeDefined || defaultRouteAcceptable:
		for _, d := range appParams.Domains {
			domain, err := cmd.findDomain(&d)
			if err != nil {
				return err
			}
			appParams.UseRandomPort = isTCP(domain)
			err = cmd.processDomainsAndBindRoutes(appParams, app, domain)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

const TCP = "tcp"

func isTCP(domain models.DomainFields) bool {
	return domain.RouterGroupType == TCP
}

func (cmd *Push) processDomainsAndBindRoutes(
	appParams models.AppParams,
	app models.Application,
	domain models.DomainFields,
) error {
	if appParams.IsHostEmpty() {
		err := cmd.createAndBindRoute(
			nil,
			appParams.UseRandomRoute,
			appParams.UseRandomPort,
			app,
			appParams.IsNoHostnameTrue(),
			domain,
			appParams.RoutePath,
		)
		if err != nil {
			return err
		}
	} else {
		for _, host := range appParams.Hosts {
			err := cmd.createAndBindRoute(
				&host,
				appParams.UseRandomRoute,
				appParams.UseRandomPort,
				app,
				appParams.IsNoHostnameTrue(),
				domain,
				appParams.RoutePath,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cmd *Push) createAndBindRoute(
	host *string,
	UseRandomRoute bool,
	UseRandomPort bool,
	app models.Application,
	noHostName bool,
	domain models.DomainFields,
	routePath *string,
) error {
	var hostname string
	if !noHostName {
		switch {
		case host != nil:
			hostname = *host
		case UseRandomPort:
			//do nothing
		case UseRandomRoute:
			hostname = hostNameForString(app.Name) + "-" + cmd.wordGenerator.Babble()
		default:
			hostname = hostNameForString(app.Name)
		}
	}

	var route models.Route
	var err error
	if routePath != nil {
		route, err = cmd.routeActor.FindOrCreateRoute(hostname, domain, *routePath, 0, UseRandomPort)
	} else {
		route, err = cmd.routeActor.FindOrCreateRoute(hostname, domain, "", 0, UseRandomPort)
	}
	if err != nil {
		return err
	}
	return cmd.routeActor.BindRoute(app, route)
}

var forbiddenHostCharRegex = regexp.MustCompile("[^a-z0-9-]")
var whitespaceRegex = regexp.MustCompile(`[\s_]+`)

func hostNameForString(name string) string {
	name = strings.ToLower(name)
	name = whitespaceRegex.ReplaceAllString(name, "-")
	name = forbiddenHostCharRegex.ReplaceAllString(name, "")
	return name
}

func (cmd *Push) findDomain(domainName *string) (models.DomainFields, error) {
	domain, err := cmd.domainRepo.FirstOrDefault(cmd.config.OrganizationFields().GUID, domainName)
	if err != nil {
		return models.DomainFields{}, err
	}

	return domain, nil
}

func (cmd *Push) bindAppToServices(services []string, app models.Application) error {
	for _, serviceName := range services {
		serviceInstance, err := cmd.serviceRepo.FindInstanceByName(serviceName)

		if err != nil {
			return errors.New(T("Could not find service {{.ServiceName}} to bind to {{.AppName}}",
				map[string]interface{}{"ServiceName": serviceName, "AppName": app.Name}))
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
		case errors.HTTPError:
			if httpErr.ErrorCode() == errors.ServiceBindingAppServiceTaken {
				err = nil
			}
		}

		if err != nil {
			return errors.New(T("Could not bind to service {{.ServiceName}}\nError: {{.Err}}",
				map[string]interface{}{"ServiceName": serviceName, "Err": err.Error()}))
		}

		cmd.ui.Ok()
	}
	return nil
}

func (cmd *Push) fetchStackGUID(appParams *models.AppParams) error {
	if appParams.StackName == nil {
		return nil
	}

	stackName := *appParams.StackName
	cmd.ui.Say(T("Using stack {{.StackName}}...",
		map[string]interface{}{"StackName": terminal.EntityNameColor(stackName)}))

	stack, err := cmd.stackRepo.FindByName(stackName)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	appParams.StackGUID = &stack.GUID
	return nil
}

func (cmd *Push) restart(app models.Application, params models.AppParams, c flags.FlagContext) error {
	if app.State != T("stopped") {
		cmd.ui.Say("")
		app, _ = cmd.appStopper.ApplicationStop(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
	}

	cmd.ui.Say("")

	if c.Bool("no-start") {
		return nil
	}

	if params.HealthCheckTimeout != nil {
		cmd.appStarter.SetStartTimeoutInSeconds(*params.HealthCheckTimeout)
	}

	_, err := cmd.appStarter.ApplicationStart(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *Push) getAppParamsFromManifest(c flags.FlagContext) ([]models.AppParams, error) {
	if c.Bool("no-manifest") {
		return []models.AppParams{}, nil
	}

	var path string
	if c.String("f") != "" {
		path = c.String("f")
	} else {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return nil, errors.New(fmt.Sprint(T("Could not determine the current working directory!"), err))
		}
	}

	m, err := cmd.manifestRepo.ReadManifest(path)

	if err != nil {
		if m.Path == "" && c.String("f") == "" {
			return []models.AppParams{}, nil
		}
		return nil, errors.New(T("Error reading manifest file:\n{{.Err}}", map[string]interface{}{"Err": err.Error()}))
	}

	apps, err := m.Applications()
	if err != nil {
		return nil, errors.New(T("Error reading manifest file:\n{{.Err}}", map[string]interface{}{"Err": err.Error()}))
	}

	cmd.ui.Say(T("Using manifest file {{.Path}}\n",
		map[string]interface{}{"Path": terminal.EntityNameColor(m.Path)}))
	return apps, nil
}

func (cmd *Push) createAppSetFromContextAndManifest(contextApp models.AppParams, manifestApps []models.AppParams) ([]models.AppParams, error) {
	var err error
	var apps []models.AppParams

	switch len(manifestApps) {
	case 0:
		if contextApp.Name == nil {
			return nil, errors.New(
				T("Incorrect Usage. The push command requires an app name. The app name can be supplied as an argument or with a manifest.yml file.") +
					"\n\n" +
					commandregistry.Commands.CommandUsage("push"),
			)
		}
		err = addApp(&apps, contextApp)
	case 1:
		err = checkCombinedDockerProperties(contextApp, manifestApps[0])
		if err != nil {
			return nil, err
		}

		manifestApps[0].Merge(&contextApp)
		err = addApp(&apps, manifestApps[0])
	default:
		selectedAppName := contextApp.Name
		contextApp.Name = nil

		if !contextApp.IsEmpty() {
			return nil, errors.New(T("Incorrect Usage. Command line flags (except -f) cannot be applied when pushing multiple apps from a manifest file."))
		}

		if selectedAppName != nil {
			var foundApp bool
			for _, appParams := range manifestApps {
				if appParams.Name != nil && *appParams.Name == *selectedAppName {
					foundApp = true
					err = addApp(&apps, appParams)
				}
			}

			if !foundApp {
				err = errors.New(T("Could not find app named '{{.AppName}}' in manifest", map[string]interface{}{"AppName": *selectedAppName}))
			}
		} else {
			for _, manifestApp := range manifestApps {
				err = addApp(&apps, manifestApp)
			}
		}
	}

	if err != nil {
		return nil, errors.New(T("Error: {{.Err}}", map[string]interface{}{"Err": err.Error()}))
	}

	return apps, nil
}

func checkCombinedDockerProperties(flagContext models.AppParams, manifestApp models.AppParams) error {
	if manifestApp.DockerUsername != nil || flagContext.DockerUsername != nil {
		if manifestApp.DockerImage == nil && flagContext.DockerImage == nil {
			return errors.New(T("'--docker-username' requires '--docker-image' to be specified"))
		}
	}

	dockerPassword := os.Getenv("CF_DOCKER_PASSWORD")
	if flagContext.DockerUsername == nil && manifestApp.DockerUsername != nil && dockerPassword == "" {
		return errors.New(T("No Docker password was provided. Please provide the password by setting the CF_DOCKER_PASSWORD environment variable."))
	}

	return nil
}

func addApp(apps *[]models.AppParams, app models.AppParams) error {
	if app.Name == nil {
		return errors.New(T("App name is a required field"))
	}

	if app.Path == nil {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		app.Path = &cwd
	}

	*apps = append(*apps, app)

	return nil
}

func (cmd *Push) getAppParamsFromContext(c flags.FlagContext) (models.AppParams, error) {
	noHostBool := c.Bool("no-hostname")
	appParams := models.AppParams{
		NoRoute:        c.Bool("no-route"),
		UseRandomRoute: c.Bool("random-route"),
		NoHostname:     &noHostBool,
	}

	if len(c.Args()) > 0 {
		appParams.Name = &c.Args()[0]
	}

	if c.String("n") != "" {
		appParams.Hosts = []string{c.String("n")}
	}

	if c.String("route-path") != "" {
		routePath := c.String("route-path")
		appParams.RoutePath = &routePath
	}

	if c.String("app-ports") != "" {
		appPortStrings := strings.Split(c.String("app-ports"), ",")
		appPorts := make([]int, len(appPortStrings))

		for i, s := range appPortStrings {
			p, err := strconv.Atoi(s)
			if err != nil {
				return models.AppParams{}, errors.New(T("Invalid app port: {{.AppPort}}\nApp port must be a number", map[string]interface{}{
					"AppPort": s,
				}))
			}
			appPorts[i] = p
		}

		appParams.AppPorts = &appPorts
	}

	if c.String("b") != "" {
		buildpack := c.String("b")
		if buildpack == "null" || buildpack == "default" {
			buildpack = ""
		}
		appParams.BuildpackURL = &buildpack
	}

	if c.String("c") != "" {
		command := c.String("c")
		if command == "null" || command == "default" {
			command = ""
		}
		appParams.Command = &command
	}

	if c.String("d") != "" {
		appParams.Domains = []string{c.String("d")}
	}

	if c.IsSet("i") {
		instances := c.Int("i")
		if instances < 1 {
			return models.AppParams{}, errors.New(T("Invalid instance count: {{.InstancesCount}}\nInstance count must be a positive integer",
				map[string]interface{}{"InstancesCount": instances}))
		}
		appParams.InstanceCount = &instances
	}

	if c.String("k") != "" {
		diskQuota, err := formatters.ToMegabytes(c.String("k"))
		if err != nil {
			return models.AppParams{}, errors.New(T("Invalid disk quota: {{.DiskQuota}}\n{{.Err}}",
				map[string]interface{}{"DiskQuota": c.String("k"), "Err": err.Error()}))
		}
		appParams.DiskQuota = &diskQuota
	}

	if c.String("m") != "" {
		memory, err := formatters.ToMegabytes(c.String("m"))
		if err != nil {
			return models.AppParams{}, errors.New(T("Invalid memory limit: {{.MemLimit}}\n{{.Err}}",
				map[string]interface{}{"MemLimit": c.String("m"), "Err": err.Error()}))
		}
		appParams.Memory = &memory
	}

	if c.String("docker-image") != "" {
		dockerImage := c.String("docker-image")
		appParams.DockerImage = &dockerImage
	}

	if c.String("docker-username") != "" {
		username := c.String("docker-username")
		appParams.DockerUsername = &username

		password := os.Getenv("CF_DOCKER_PASSWORD")
		if password != "" {
			cmd.ui.Say(T("Using docker repository password from environment variable CF_DOCKER_PASSWORD."))
		} else {
			cmd.ui.Say(T("Environment variable CF_DOCKER_PASSWORD not set."))
			password = cmd.ui.AskForPassword("Docker password")
			if password == "" {
				return models.AppParams{}, errors.New(T("Please provide a password."))
			}
		}
		appParams.DockerPassword = &password
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
			return models.AppParams{}, fmt.Errorf("Error: %s", fmt.Errorf(T("Invalid timeout param: {{.Timeout}}\n{{.Err}}",
				map[string]interface{}{"Timeout": c.String("t"), "Err": err.Error()})))
		}

		appParams.HealthCheckTimeout = &timeout
	}

	healthCheckType := c.String("u")
	switch healthCheckType {
	case "":
		// do nothing
	case "http", "none", "port", "process":
		appParams.HealthCheckType = &healthCheckType
	default:
		return models.AppParams{}, fmt.Errorf("Error: %s", fmt.Errorf(T("Invalid health-check-type param: {{.healthCheckType}}",
			map[string]interface{}{"healthCheckType": healthCheckType})))
	}

	return appParams, nil
}

func (cmd Push) ValidateContextAndAppParams(appsFromManifest []models.AppParams, appFromContext models.AppParams) error {
	if appFromContext.NoHostname != nil && *appFromContext.NoHostname {
		for _, app := range appsFromManifest {
			if app.Routes != nil {
				return errors.New(T("Option '--no-hostname' cannot be used with an app manifest containing the 'routes' attribute"))
			}
		}
	}

	return nil
}

func (cmd *Push) uploadApp(appGUID, appDir, appDirOrZipFile string, localFiles []models.AppFileFields) error {
	uploadDir, err := ioutil.TempDir("", "apps")
	if err != nil {
		return err
	}

	remoteFiles, hasFileToUpload, err := cmd.actor.GatherFiles(localFiles, appDir, uploadDir, true)

	if httpError, isHTTPError := err.(errors.HTTPError); isHTTPError && httpError.StatusCode() == 504 {
		cmd.ui.Warn("Resource matching API timed out; pushing all app files.")
		remoteFiles, hasFileToUpload, err = cmd.actor.GatherFiles(localFiles, appDir, uploadDir, false)
	}

	if err != nil {
		return err
	}

	zipFile, err := ioutil.TempFile("", "uploads")
	if err != nil {
		return err
	}
	defer func() {
		zipFile.Close()
		os.Remove(zipFile.Name())
	}()

	if hasFileToUpload {
		err = cmd.zipper.Zip(uploadDir, zipFile)
		if err != nil {
			if emptyDirErr, ok := err.(*errors.EmptyDirError); ok {
				return emptyDirErr
			}
			return fmt.Errorf("%s: %s", T("Error zipping application"), err.Error())
		}

		var zipFileSize int64
		zipFileSize, err = cmd.zipper.GetZipSize(zipFile)
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
	}

	err = os.RemoveAll(uploadDir)
	if err != nil {
		return err
	}

	return cmd.actor.UploadApp(appGUID, zipFile, remoteFiles)
}
