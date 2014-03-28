package api

import (
	"cf/errors"
	"cf/models"
)

type FakeRouteRepository struct {
	FindByHostAndDomainHost     string
	FindByHostAndDomainDomain   string
	FindByHostAndDomainRoute    models.Route
	FindByHostAndDomainErr      bool
	FindByHostAndDomainNotFound bool

	CreatedHost       string
	CreatedDomainGuid string
	CreatedRoute      models.Route

	CreateInSpaceHost         string
	CreateInSpaceDomainGuid   string
	CreateInSpaceSpaceGuid    string
	CreateInSpaceCreatedRoute models.Route
	CreateInSpaceErr          bool

	BindErr        error
	BoundRouteGuid string
	BoundAppGuid   string

	UnboundRouteGuid string
	UnboundAppGuid   string

	ListErr bool
	Routes  []models.Route

	DeletedRouteGuids []string
	DeleteErr         error
}

func (repo *FakeRouteRepository) ListRoutes(cb func(models.Route) bool) (apiErr error) {
	if repo.ListErr {
		return errors.New("WHOOPSIE")
	}

	for _, route := range repo.Routes {
		if !cb(route) {
			break
		}
	}
	return
}

func (repo *FakeRouteRepository) FindByHostAndDomain(host, domain string) (route models.Route, apiErr error) {
	repo.FindByHostAndDomainHost = host
	repo.FindByHostAndDomainDomain = domain

	if repo.FindByHostAndDomainErr {
		apiErr = errors.NewModelNotFoundError("Route", host)
	}

	if repo.FindByHostAndDomainNotFound {
		apiErr = errors.NewModelNotFoundError("Org", host+"."+domain)
	}

	route = repo.FindByHostAndDomainRoute
	return
}

func (repo *FakeRouteRepository) Create(host, domainGuid string) (createdRoute models.Route, apiErr error) {
	repo.CreatedHost = host
	repo.CreatedDomainGuid = domainGuid

	createdRoute.Guid = host + "-route-guid"

	return
}

func (repo *FakeRouteRepository) CreateInSpace(host, domainGuid, spaceGuid string) (createdRoute models.Route, apiErr error) {
	repo.CreateInSpaceHost = host
	repo.CreateInSpaceDomainGuid = domainGuid
	repo.CreateInSpaceSpaceGuid = spaceGuid

	if repo.CreateInSpaceErr {
		apiErr = errors.New("Error")
	} else {
		createdRoute = repo.CreateInSpaceCreatedRoute
	}

	return
}

func (repo *FakeRouteRepository) Bind(routeGuid, appGuid string) (apiErr error) {
	repo.BoundRouteGuid = routeGuid
	repo.BoundAppGuid = appGuid
	return repo.BindErr
}

func (repo *FakeRouteRepository) Unbind(routeGuid, appGuid string) (apiErr error) {
	repo.UnboundRouteGuid = routeGuid
	repo.UnboundAppGuid = appGuid
	return
}

func (repo *FakeRouteRepository) Delete(routeGuid string) (apiErr error) {
	repo.DeletedRouteGuids = append(repo.DeletedRouteGuids, routeGuid)
	apiErr = repo.DeleteErr
	return
}
