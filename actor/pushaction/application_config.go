package pushaction

import (
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	"code.cloudfoundry.org/cli/actor/v2action"
	log "github.com/Sirupsen/logrus"
)

type ApplicationConfig struct {
	CurrentApplication v2action.Application
	DesiredApplication v2action.Application

	CurrentRoutes []v2action.Route
	DesiredRoutes []v2action.Route

	TargetedSpaceGUID string
	Path              string
}

func (actor Actor) ConvertToApplicationConfig(orgGUID string, spaceGUID string, apps []manifest.Application) ([]ApplicationConfig, Warnings, error) {
	var configs []ApplicationConfig
	var warnings Warnings

	log.Infof("iterating through %d app configuration(s)", len(apps))
	for _, app := range apps {
		config := ApplicationConfig{
			TargetedSpaceGUID: spaceGUID,
			Path:              app.Path,
		}

		log.Infoln("searching for app", app.Name)
		appExists, foundApp, v2Warnings, err := actor.FindOrReturnParialApp(app.Name, spaceGUID)
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

		defaultRoute, routeWarnings, err := actor.GetRouteWithDefaultDomain(app.Name, orgGUID, spaceGUID)
		warnings = append(warnings, routeWarnings...)
		if err != nil {
			log.Errorln("getting default route:", err)
			return nil, warnings, err
		}
		config.DesiredRoutes = []v2action.Route{defaultRoute}

		configs = append(configs, config)
	}

	return configs, warnings, nil
}

func (actor Actor) FindOrReturnParialApp(appName string, spaceGUID string) (bool, v2action.Application, v2action.Warnings, error) {
	foundApp, v2Warnings, err := actor.V2Actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if _, ok := err.(v2action.ApplicationNotFoundError); ok {
		log.Warnf("unable to find app %s in current space (GUID: %s)", appName, spaceGUID)
		return false, v2action.Application{}, v2Warnings, nil
	}
	return true, foundApp, v2Warnings, err
}
