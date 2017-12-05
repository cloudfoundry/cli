package pushaction

import (
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/util/manifest"
	log "github.com/sirupsen/logrus"
)

type ApplicationConfig struct {
	CurrentApplication Application
	DesiredApplication Application

	CurrentRoutes []v2action.Route
	DesiredRoutes []v2action.Route
	NoRoute       bool

	CurrentServices map[string]v2action.ServiceInstance
	DesiredServices map[string]v2action.ServiceInstance

	AllResources       []v2action.Resource
	MatchedResources   []v2action.Resource
	UnmatchedResources []v2action.Resource
	Archive            bool
	Path               string

	TargetedSpaceGUID string
}

func (config ApplicationConfig) CreatingApplication() bool {
	return config.CurrentApplication.GUID == ""
}

func (config ApplicationConfig) UpdatingApplication() bool {
	return !config.CreatingApplication()
}

func (actor Actor) ConvertToApplicationConfigs(orgGUID string, spaceGUID string, noStart bool, apps []manifest.Application) ([]ApplicationConfig, Warnings, error) {
	var configs []ApplicationConfig
	var warnings Warnings

	log.Infof("iterating through %d app configuration(s)", len(apps))
	for _, app := range apps {
		absPath, err := filepath.EvalSymlinks(app.Path)
		if err != nil {
			return nil, nil, err
		}

		config := ApplicationConfig{
			TargetedSpaceGUID: spaceGUID,
			Path:              absPath,
			NoRoute:           app.NoRoute,
		}

		log.Infoln("searching for app", app.Name)
		found, constructedApp, v2Warnings, err := actor.FindOrReturnPartialApp(app.Name, spaceGUID)
		warnings = append(warnings, v2Warnings...)
		if err != nil {
			log.Errorln("app lookup:", err)
			return nil, warnings, err
		}

		if found {
			var configWarnings v2action.Warnings
			config, configWarnings, err = actor.configureExistingApp(config, app, constructedApp)
			warnings = append(warnings, configWarnings...)
			if err != nil {
				log.Errorln("configuring existing app:", err)
				return nil, warnings, err
			}
		} else {
			log.Debug("using empty app as base")
			config.DesiredApplication = constructedApp
		}
		config.DesiredApplication = actor.overrideApplicationProperties(config.DesiredApplication, app, noStart)

		var stackWarnings Warnings
		config.DesiredApplication, stackWarnings, err = actor.overrideStack(config.DesiredApplication, app)
		warnings = append(warnings, stackWarnings...)
		if err != nil {
			return nil, warnings, err
		}
		log.Debugln("post overriding config:", config.DesiredApplication)

		var serviceWarnings Warnings
		config.DesiredServices, serviceWarnings, err = actor.getDesiredServices(config.CurrentServices, app.Services, spaceGUID)
		warnings = append(warnings, serviceWarnings...)
		if err != nil {
			log.Errorln("getting services:", err)
			return nil, warnings, err
		}

		if !config.NoRoute {
			var routeWarnings Warnings
			config, routeWarnings, err = actor.configureRoutes(app, orgGUID, spaceGUID, config)
			warnings = append(warnings, routeWarnings...)
			if err != nil {
				log.Errorln("determining routes:", err)
				return nil, warnings, err
			}
		}

		if app.DockerImage == "" {
			config, err = actor.configureResources(config)
			if err != nil {
				log.Errorln("configuring resources", err)
				return nil, warnings, err
			}
		}

		configs = append(configs, config)
	}

	return configs, warnings, nil
}

func (actor Actor) configureRoutes(manifestApp manifest.Application, orgGUID string, spaceGUID string, config ApplicationConfig) (ApplicationConfig, Warnings, error) {
	switch {
	case len(manifestApp.Routes) > 0: // Routes in manifest mutually exclusive with all other route related fields
		routes, warnings, err := actor.CalculateRoutes(manifestApp.Routes, orgGUID, spaceGUID, config.CurrentRoutes)
		config.DesiredRoutes = routes
		return config, warnings, err
	case manifestApp.RandomRoute && len(config.CurrentRoutes) > 0:
		config.DesiredRoutes = config.CurrentRoutes
		return config, nil, nil
	case manifestApp.RandomRoute:
		// append random route to current route (becomes desired route)
		randomRoute, warnings, err := actor.GenerateRandomRoute(manifestApp, spaceGUID, orgGUID)
		config.DesiredRoutes = append(config.CurrentRoutes, randomRoute)
		return config, warnings, err
	default:
		desiredRoute, warnings, err := actor.GetGeneratedRoute(manifestApp, orgGUID, spaceGUID, config.CurrentRoutes)
		if err != nil {
			log.Errorln("getting default route:", err)
			return config, warnings, err
		}
		config.DesiredRoutes = append(config.CurrentRoutes, desiredRoute)
		return config, warnings, nil
	}
}

func (actor Actor) getDesiredServices(currentServices map[string]v2action.ServiceInstance, requestedServices []string, spaceGUID string) (map[string]v2action.ServiceInstance, Warnings, error) {
	var warnings Warnings

	desiredServices := map[string]v2action.ServiceInstance{}
	for name, serviceInstance := range currentServices {
		log.Debugln("adding bound service:", name)
		desiredServices[name] = serviceInstance
	}

	for _, serviceName := range requestedServices {
		if _, ok := desiredServices[serviceName]; !ok {
			log.Debugln("adding requested service:", serviceName)
			serviceInstance, serviceWarnings, err := actor.V2Actor.GetServiceInstanceByNameAndSpace(serviceName, spaceGUID)
			warnings = append(warnings, serviceWarnings...)
			if err != nil {
				return nil, warnings, err
			}

			desiredServices[serviceName] = serviceInstance
		}
	}
	return desiredServices, warnings, nil
}

func (actor Actor) configureExistingApp(config ApplicationConfig, app manifest.Application, foundApp Application) (ApplicationConfig, v2action.Warnings, error) {
	log.Debugln("found app:", foundApp)
	config.CurrentApplication = foundApp
	config.DesiredApplication = foundApp

	log.Info("looking up application routes")
	routes, warnings, err := actor.V2Actor.GetApplicationRoutes(foundApp.GUID)
	if err != nil {
		log.Errorln("existing routes lookup:", err)
		return config, warnings, err
	}

	serviceInstances, serviceWarnings, err := actor.V2Actor.GetServiceInstancesByApplication(foundApp.GUID)
	warnings = append(warnings, serviceWarnings...)
	if err != nil {
		log.Errorln("existing services lookup:", err)
		return config, warnings, err
	}

	nameToService := map[string]v2action.ServiceInstance{}
	for _, serviceInstance := range serviceInstances {
		nameToService[serviceInstance.Name] = serviceInstance
	}

	config.CurrentRoutes = routes
	config.CurrentServices = nameToService
	return config, warnings, nil
}

func (actor Actor) configureResources(config ApplicationConfig) (ApplicationConfig, error) {
	info, err := os.Stat(config.Path)
	if err != nil {
		return config, err
	}

	var resources []sharedaction.Resource
	if info.IsDir() {
		log.WithField("path_to_resources", config.Path).Info("determine directory resources to zip")
		resources, err = actor.SharedActor.GatherDirectoryResources(config.Path)
	} else {
		config.Archive = true
		log.WithField("path_to_resources", config.Path).Info("determine archive resources to zip")
		resources, err = actor.SharedActor.GatherArchiveResources(config.Path)
	}
	if err != nil {
		return config, err
	}
	config.AllResources = actor.ConvertSharedResourcesToV2Resources(resources)

	log.WithField("number_of_files", len(resources)).Debug("completed file scan")

	return config, nil
}

func (Actor) overrideApplicationProperties(application Application, manifest manifest.Application, noStart bool) Application {
	if manifest.Buildpack.IsSet {
		application.Buildpack = manifest.Buildpack
	}
	if manifest.Command.IsSet {
		application.Command = manifest.Command
	}
	if manifest.DockerImage != "" {
		application.DockerImage = manifest.DockerImage
		if manifest.DockerUsername != "" {
			application.DockerCredentials.Username = manifest.DockerUsername
			application.DockerCredentials.Password = manifest.DockerPassword
		}
	}

	if manifest.DiskQuota.IsSet {
		application.DiskQuota = manifest.DiskQuota
	}

	if manifest.Memory.IsSet {
		application.Memory = manifest.Memory
	}

	if manifest.HealthCheckTimeout != 0 {
		application.HealthCheckTimeout = manifest.HealthCheckTimeout
	}

	if manifest.HealthCheckType != "" {
		application.HealthCheckType = constant.ApplicationHealthCheckType(manifest.HealthCheckType)
		application.HealthCheckHTTPEndpoint = manifest.HealthCheckHTTPEndpoint

		if application.HealthCheckType == constant.ApplicationHealthCheckHTTP && application.HealthCheckHTTPEndpoint == "" {
			application.HealthCheckHTTPEndpoint = "/"
		}
	}

	if manifest.Instances.IsSet {
		application.Instances = manifest.Instances
	}

	if noStart {
		application.State = ccv2.ApplicationStopped
	}

	if len(manifest.EnvironmentVariables) > 0 {
		if application.EnvironmentVariables == nil {
			application.EnvironmentVariables = manifest.EnvironmentVariables
		} else {
			env := map[string]string{}
			for key, value := range application.EnvironmentVariables {
				env[key] = value
			}
			for key, value := range manifest.EnvironmentVariables {
				env[key] = value
			}
			application.EnvironmentVariables = env
		}
	}

	log.Debugln("post application override:", application)

	return application
}

func (actor Actor) overrideStack(application Application, manifest manifest.Application) (Application, Warnings, error) {
	if manifest.StackName == "" {
		return application, nil, nil
	}
	stack, warnings, err := actor.V2Actor.GetStackByName(manifest.StackName)
	application.SetStack(stack)
	return application, Warnings(warnings), err
}
