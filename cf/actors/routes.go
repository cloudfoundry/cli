package actors

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type RouteActor struct {
	ui        terminal.UI
	routeRepo api.RouteRepository
}

func NewRouteActor(ui terminal.UI, routeRepo api.RouteRepository) RouteActor {
	return RouteActor{ui: ui, routeRepo: routeRepo}
}

func (routeActor RouteActor) FindOrCreateRoute(hostname string, domain models.DomainFields) (route models.Route) {
	route, apiErr := routeActor.routeRepo.FindByHostAndDomain(hostname, domain)

	switch apiErr.(type) {
	case nil:
		routeActor.ui.Say("Using route %s", terminal.EntityNameColor(route.URL()))
	case *errors.ModelNotFoundError:
		routeActor.ui.Say("Creating route %s...", terminal.EntityNameColor(domain.UrlForHost(hostname)))

		route, apiErr = routeActor.routeRepo.Create(hostname, domain)
		if apiErr != nil {
			routeActor.ui.Failed(apiErr.Error())
		}

		routeActor.ui.Ok()
		routeActor.ui.Say("")
	default:
		routeActor.ui.Failed(apiErr.Error())
	}

	return
}

func (routeActor RouteActor) BindRoute(app models.Application, route models.Route) {
	if !app.HasRoute(route) {
		routeActor.ui.Say("Binding %s to %s...", terminal.EntityNameColor(route.URL()), terminal.EntityNameColor(app.Name))

		apiErr := routeActor.routeRepo.Bind(route.Guid, app.Guid)
		switch apiErr := apiErr.(type) {
		case nil:
			routeActor.ui.Ok()
			routeActor.ui.Say("")
			return
		case errors.HttpError:
			if apiErr.ErrorCode() == errors.INVALID_RELATION {
				routeActor.ui.Failed("The route %s is already in use.\nTIP: Change the hostname with -n HOSTNAME or use --random-route to generate a new route and then push again.", route.URL())
			}
		}
		routeActor.ui.Failed(apiErr.Error())
	}
}

func (routeActor RouteActor) UnbindAll(app models.Application) {
	for _, route := range app.Routes {
		routeActor.ui.Say("Removing route %s...", terminal.EntityNameColor(route.URL()))
		routeActor.routeRepo.Unbind(route.Guid, app.Guid)
	}
}
