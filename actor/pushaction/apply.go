package pushaction

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	log "github.com/Sirupsen/logrus"
)

func (actor Actor) Apply(config ApplicationConfig) (<-chan Event, <-chan Warnings, <-chan error) {
	eventStream := make(chan Event)
	warningsStream := make(chan Warnings)
	errorStream := make(chan error)

	go func() {
		log.Debug("starting apply go routine")
		defer close(eventStream)
		defer close(warningsStream)
		defer close(errorStream)

		if config.DesiredApplication.GUID != "" {
			log.Debugf("updating application: %#v", config.DesiredApplication)
			app, warnings, err := actor.V2Actor.UpdateApplication(config.DesiredApplication)
			warningsStream <- Warnings(warnings)
			if err != nil {
				log.Errorln("updating application:", err)
				errorStream <- err
				return
			}
			config.DesiredApplication = app
			eventStream <- ApplicationUpdated
		} else {
			log.Debugf("creating application: %#v", config.DesiredApplication)
			app, warnings, err := actor.V2Actor.CreateApplication(config.DesiredApplication)
			warningsStream <- Warnings(warnings)
			if err != nil {
				log.Errorln("creating application:", err)
				errorStream <- err
				return
			}
			config.DesiredApplication = app
			eventStream <- ApplicationCreated
		}
		log.Debugf("desired application: %#v", config.DesiredApplication)

		log.Info("creating routes")
		var createdRoutes []v2action.Route
		var createdRoutesMessage bool
		for _, route := range config.DesiredRoutes {
			if route.GUID == "" {
				log.Debugf("creating route: %#v", route)
				createdRoute, warnings, err := actor.V2Actor.CreateRoute(route, false)
				warningsStream <- Warnings(warnings)
				if err != nil {
					log.Errorln("creating route:", err)
					errorStream <- err
					return
				}
				createdRoutes = append(createdRoutes, createdRoute)
				createdRoutesMessage = true
			} else {
				log.Debugf("route %s already exists, skipping creation", route)
				createdRoutes = append(createdRoutes, route)
			}
		}
		config.DesiredRoutes = createdRoutes

		if createdRoutesMessage {
			log.Debugf("updated desired routes: %#v", config.DesiredRoutes)
			eventStream <- RouteCreated
		}

		log.Info("binding routes")
		var boundRoutesMessage bool
		for _, route := range config.DesiredRoutes {
			if !actor.routeInList(route, config.CurrentRoutes) {
				log.Debugf("binding route: %#v", route)
				warnings, err := actor.bindRouteToApp(route, config.DesiredApplication.GUID)
				warningsStream <- Warnings(warnings)
				if err != nil {
					log.Errorln("binding route:", err)
					errorStream <- err
					return
				}
				boundRoutesMessage = true
			} else {
				log.Debugf("route %s already bound to app", route)
			}
		}
		log.Debug("binding routes complete")
		config.CurrentRoutes = config.DesiredRoutes

		if boundRoutesMessage {
			eventStream <- RouteBound
		}

		log.Debug("completed apply")
		eventStream <- Complete
	}()

	return eventStream, warningsStream, errorStream
}
func (actor Actor) bindRouteToApp(route v2action.Route, appGUID string) (v2action.Warnings, error) {
	warnings, err := actor.V2Actor.BindRouteToApplication(route.GUID, appGUID)
	if _, ok := err.(v2action.RouteInDifferentSpaceError); ok {
		return warnings, v2action.RouteInDifferentSpaceError{Route: route.String()}
	}
	return warnings, err
}
