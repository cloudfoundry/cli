package pushaction

import (
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	"code.cloudfoundry.org/cli/actor/v2action"
	log "github.com/sirupsen/logrus"
)

type ApplicationConfig struct {
	CurrentApplication v2action.Application
	DesiredApplication v2action.Application

	CurrentRoutes []v2action.Route
	DesiredRoutes []v2action.Route

	AllResources []v2action.Resource

	TargetedSpaceGUID string
	Path              string
}

func (config ApplicationConfig) CreatingApplication() bool {
	return config.CurrentApplication.GUID == ""
}

func (config ApplicationConfig) UpdatingApplication() bool {
	return !config.CreatingApplication()
}

func (actor Actor) ConvertToApplicationConfigs(orgGUID string, spaceGUID string, apps []manifest.Application) ([]ApplicationConfig, Warnings, error) {
	var configs []ApplicationConfig
	var warnings Warnings

	log.Infof("iterating through %d app configuration(s)", len(apps))
	for _, app := range apps {
		config := ApplicationConfig{
			TargetedSpaceGUID: spaceGUID,
			Path:              app.Path,
		}

		log.Infoln("searching for app", app.Name)
		appExists, foundApp, v2Warnings, err := actor.FindOrReturnPartialApp(app.Name, spaceGUID)
		warnings = append(warnings, v2Warnings...)
		if err != nil {
			log.Errorln("app lookup:", err)
			return nil, warnings, err
		}

		if appExists {
			log.Debugf("found app: %#v", foundApp)
			config.CurrentApplication = foundApp
			config.DesiredApplication = foundApp

			log.Info("looking up application routes")
			var routes []v2action.Route
			var routeWarnings v2action.Warnings
			routes, routeWarnings, err = actor.V2Actor.GetApplicationRoutes(foundApp.GUID)
			warnings = append(warnings, routeWarnings...)
			if err != nil {
				log.Errorln("existing routes lookup:", err)
				return nil, warnings, err
			}
			config.CurrentRoutes = routes
		} else {
			log.Debug("using empty app as base")
			config.DesiredApplication.Name = app.Name
			config.DesiredApplication.SpaceGUID = spaceGUID
		}

		defaultRoute, routeWarnings, err := actor.GetRouteWithDefaultDomain(app.Name, orgGUID, spaceGUID, config.CurrentRoutes)
		warnings = append(warnings, routeWarnings...)
		if err != nil {
			log.Errorln("getting default route:", err)
			return nil, warnings, err
		}

		// TODO: when working with all of routes, append to current route
		config.DesiredRoutes = []v2action.Route{defaultRoute}

		if app.DockerImage == "" {
			log.WithField("path_to_resources", app.Path).Info("determine resources to zip")
			resources, err := actor.V2Actor.GatherDirectoryResources(app.Path)
			if err != nil {
				return nil, warnings, err
			}
			config.AllResources = resources
			log.WithField("number_of_files", len(resources)).Debug("completed file scan")
		} else {
			config.DesiredApplication.DockerImage = app.DockerImage
		}

		configs = append(configs, config)
	}

	return configs, warnings, nil
}

func (actor Actor) FindOrReturnPartialApp(appName string, spaceGUID string) (bool, v2action.Application, v2action.Warnings, error) {
	foundApp, v2Warnings, err := actor.V2Actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if _, ok := err.(v2action.ApplicationNotFoundError); ok {
		log.Warnf("unable to find app %s in current space (GUID: %s)", appName, spaceGUID)
		return false, v2action.Application{}, v2Warnings, nil
	}
	return true, foundApp, v2Warnings, err
}
