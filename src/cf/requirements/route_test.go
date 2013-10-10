package requirements_test

import (
	"cf"
	. "cf/requirements"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testterm "testhelpers/terminal"
	"testing"
)

func TestRouteReqExecute(t *testing.T) {
	route := cf.Route{
		Host:   "my-route",
		Domain: cf.Domain{Name: "example.com", Guid: "my-domain-guid"},
		Guid:   "my-route-guid",
	}
	routeRepo := &testapi.FakeRouteRepository{FindByHostAndDomainRoute: route}
	ui := new(testterm.FakeUI)

	routeReq := NewRouteRequirement("host", "example.com", ui, routeRepo)
	success := routeReq.Execute()

	assert.True(t, success)
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "host")
	assert.Equal(t, routeRepo.FindByHostAndDomainDomain, "example.com")
	assert.Equal(t, routeReq.GetRoute(), route)
}

func TestRouteReqWhenRouteDoesNotExist(t *testing.T) {
	routeRepo := &testapi.FakeRouteRepository{FindByHostAndDomainNotFound: true}
	ui := new(testterm.FakeUI)

	routeReq := NewRouteRequirement("host", "example.com", ui, routeRepo)
	success := routeReq.Execute()

	assert.False(t, success)
}

func TestRouteReqOnError(t *testing.T) {
	routeRepo := &testapi.FakeRouteRepository{FindByHostAndDomainErr: true}
	ui := new(testterm.FakeUI)

	routeReq := NewRouteRequirement("host", "example.com", ui, routeRepo)
	success := routeReq.Execute()

	assert.False(t, success)
}
