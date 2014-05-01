package api

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
)

type FakeRouteRepository struct {
	FindByHostAndDomainCalledWith struct {
		Host   string
		Domain models.DomainFields
	}

	FindByHostAndDomainReturns struct {
		Route models.Route
		Error error
	}

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

func (repo *FakeRouteRepository) FindByHostAndDomain(host string, domain models.DomainFields) (route models.Route, apiErr error) {
	repo.FindByHostAndDomainCalledWith.Host = host
	repo.FindByHostAndDomainCalledWith.Domain = domain

	if repo.FindByHostAndDomainReturns.Error != nil {
		apiErr = repo.FindByHostAndDomainReturns.Error
	}

	route = repo.FindByHostAndDomainReturns.Route
	return
}

func (repo *FakeRouteRepository) Create(host string, domain models.DomainFields) (createdRoute models.Route, apiErr error) {
	repo.CreatedHost = host
	repo.CreatedDomainGuid = domain.Guid

	createdRoute.Guid = host + "-route-guid"
	createdRoute.Domain = domain
	createdRoute.Host = host

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
