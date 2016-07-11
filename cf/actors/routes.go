package actors

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

//go:generate counterfeiter . RouteActor

type RouteActor interface {
	CreateRandomTCPRoute(domain models.DomainFields) (models.Route, error)
	FindOrCreateRoute(hostname string, domain models.DomainFields, path string, useRandomPort bool) (models.Route, error)
	BindRoute(app models.Application, route models.Route) error
	UnbindAll(app models.Application) error
	FindDomain(routeName string) (string, models.DomainFields, error)
}

type routeActor struct {
	ui         terminal.UI
	routeRepo  api.RouteRepository
	domainRepo api.DomainRepository
}

func NewRouteActor(ui terminal.UI, routeRepo api.RouteRepository, domainRepo api.DomainRepository) routeActor {
	return routeActor{
		ui:         ui,
		routeRepo:  routeRepo,
		domainRepo: domainRepo,
	}
}

func (routeActor routeActor) CreateRandomTCPRoute(domain models.DomainFields) (models.Route, error) {
	routeActor.ui.Say(T("Creating random route for {{.Domain}}", map[string]interface{}{
		"Domain": terminal.EntityNameColor(domain.Name),
	}) + "...")

	route, err := routeActor.routeRepo.Create("", domain, "", true)
	if err != nil {
		return models.Route{}, err
	}

	return route, nil
}

func (routeActor routeActor) FindOrCreateRoute(hostname string, domain models.DomainFields, path string, useRandomPort bool) (models.Route, error) {
	var port int
	route, err := routeActor.routeRepo.Find(hostname, domain, path, port)

	switch err.(type) {
	case nil:
		routeActor.ui.Say(
			T("Using route {{.RouteURL}}",
				map[string]interface{}{
					"RouteURL": terminal.EntityNameColor(route.URL()),
				}),
		)
	case *errors.ModelNotFoundError:
		if useRandomPort {
			route, err = routeActor.CreateRandomTCPRoute(domain)
		} else {
			routeActor.ui.Say(
				T("Creating route {{.Hostname}}...",
					map[string]interface{}{
						"Hostname": terminal.EntityNameColor(domain.URLForHostAndPath(hostname, path, port)),
					}),
			)

			route, err = routeActor.routeRepo.Create(hostname, domain, path, useRandomPort)
		}

		routeActor.ui.Ok()
		routeActor.ui.Say("")
	}

	return route, err
}

func (routeActor routeActor) BindRoute(app models.Application, route models.Route) error {
	if !app.HasRoute(route) {
		routeActor.ui.Say(T(
			"Binding {{.URL}} to {{.AppName}}...",
			map[string]interface{}{
				"URL":     terminal.EntityNameColor(route.URL()),
				"AppName": terminal.EntityNameColor(app.Name),
			}),
		)

		err := routeActor.routeRepo.Bind(route.GUID, app.GUID)
		switch err := err.(type) {
		case nil:
			routeActor.ui.Ok()
			routeActor.ui.Say("")
			return nil
		case errors.HTTPError:
			if err.ErrorCode() == errors.InvalidRelation {
				return errors.New(T(
					"The route {{.URL}} is already in use.\nTIP: Change the hostname with -n HOSTNAME or use --random-route to generate a new route and then push again.",
					map[string]interface{}{
						"URL": route.URL(),
					}),
				)
			}
		}
		return err
	}
	return nil
}

func (routeActor routeActor) UnbindAll(app models.Application) error {
	for _, route := range app.Routes {
		routeActor.ui.Say(T(
			"Removing route {{.URL}}...",
			map[string]interface{}{
				"URL": terminal.EntityNameColor(route.URL()),
			}),
		)
		err := routeActor.routeRepo.Unbind(route.GUID, app.GUID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (routeActor routeActor) FindDomain(routeName string) (string, models.DomainFields, error) {
	domain, err := routeActor.domainRepo.FindPrivateByName(routeName)
	found, err := validateFoundDomain(domain, err)
	if err != nil {
		return "", models.DomainFields{}, err
	}
	if found {
		return "", domain, nil
	}

	domain, err = routeActor.domainRepo.FindSharedByName(routeName)
	found, err = validateFoundDomain(domain, err)
	if err != nil {
		return "", models.DomainFields{}, err
	}
	if found {
		return "", domain, nil
	}

	routeParts := strings.Split(routeName, ".")
	domain, err = routeActor.domainRepo.FindSharedByName(strings.Join(routeParts[1:], "."))
	found, err = validateFoundDomain(domain, err)
	if err != nil {
		return "", models.DomainFields{}, err
	}
	if found {
		return routeParts[0], domain, nil
	}

	return "", models.DomainFields{}, fmt.Errorf(T(
		"The route {{.RouteName}} did not match any existing domains.",
		map[string]interface{}{
			"RouteName": routeName,
		},
	))
}

func validateFoundDomain(domain models.DomainFields, err error) (bool, error) {
	switch err.(type) {
	case *errors.ModelNotFoundError:
		return false, nil
	case nil:
		if err = assertHTTPDomain(domain); err != nil {
			return false, err
		}
		return true, nil
	default:
		return false, err
	}
}

func assertHTTPDomain(domain models.DomainFields) error {
	if domain.RouterGroupType != "" {
		return fmt.Errorf(T(
			"The domain {{.DomainName}} is not an HTTP domain, unable to map route.",
			map[string]interface{}{
				"DomainName": domain.Name,
			},
		))
	}
	return nil
}
