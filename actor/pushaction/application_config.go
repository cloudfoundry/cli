package pushaction

import (
	"os"

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
	Archive      bool
	Path         string

	TargetedSpaceGUID string
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
			var configWarnings v2action.Warnings
			config, configWarnings, err = actor.configureExistingApp(config, app, foundApp)
			warnings = append(warnings, configWarnings...)
			if err != nil {
				log.Errorln("configuring existing app:", err)
				return nil, warnings, err
			}
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

		config, err = actor.configureResources(config, app)
		if err != nil {
			log.Errorln("configuring resources", err)
			return nil, warnings, err
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

func (actor Actor) configureExistingApp(config ApplicationConfig, app manifest.Application, foundApp v2action.Application) (ApplicationConfig, v2action.Warnings, error) {
	log.Debugf("found app: %#v", foundApp)
	config.CurrentApplication = foundApp
	config.DesiredApplication = foundApp

	log.Info("looking up application routes")
	routes, warnings, err := actor.V2Actor.GetApplicationRoutes(foundApp.GUID)
	if err != nil {
		log.Errorln("existing routes lookup:", err)
		return config, warnings, err
	}
	config.CurrentRoutes = routes
	return config, warnings, nil
}

func (actor Actor) configureResources(config ApplicationConfig, app manifest.Application) (ApplicationConfig, error) {
	if app.DockerImage == "" {
		info, err := os.Stat(app.Path)
		if err != nil {
			return config, err
		}

		var resources []v2action.Resource
		if info.IsDir() {
			log.WithField("path_to_resources", app.Path).Info("determine directory resources to zip")
			resources, err = actor.V2Actor.GatherDirectoryResources(app.Path)
		} else {
			config.Archive = true
			log.WithField("path_to_resources", app.Path).Info("determine archive resources to zip")
			resources, err = actor.V2Actor.GatherArchiveResources(app.Path)
		}
		if err != nil {
			return config, err
		}
		config.AllResources = resources
		log.WithField("number_of_files", len(resources)).Debug("completed file scan")
	} else {
		config.DesiredApplication.DockerImage = app.DockerImage
	}

	return config, nil
}
