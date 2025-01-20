package routefakes

import (
	"sync"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/route"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
)

type OldFakeRouteCreator struct {
	CreateRouteStub        func(hostName string, path string, port int, randomPort bool, domain models.DomainFields, space models.SpaceFields) (route models.Route, apiErr error)
	createRouteMutex       sync.RWMutex
	createRouteArgsForCall []struct {
		hostName   string
		path       string
		port       int
		randomPort bool
		domain     models.DomainFields
		space      models.SpaceFields
	}
	createRouteReturns struct {
		result1 models.Route
		result2 error
	}
}

func (fake *OldFakeRouteCreator) CreateRoute(hostName string, path string, port int, randomPort bool, domain models.DomainFields, space models.SpaceFields) (route models.Route, apiErr error) {
	fake.createRouteMutex.Lock()
	fake.createRouteArgsForCall = append(fake.createRouteArgsForCall, struct {
		hostName   string
		path       string
		port       int
		randomPort bool
		domain     models.DomainFields
		space      models.SpaceFields
	}{hostName, path, port, randomPort, domain, space})
	fake.createRouteMutex.Unlock()
	if fake.CreateRouteStub != nil {
		return fake.CreateRouteStub(hostName, path, port, randomPort, domain, space)
	} else {
		return fake.createRouteReturns.result1, fake.createRouteReturns.result2
	}
}

func (fake *OldFakeRouteCreator) CreateRouteCallCount() int {
	fake.createRouteMutex.RLock()
	defer fake.createRouteMutex.RUnlock()
	return len(fake.createRouteArgsForCall)
}

func (fake *OldFakeRouteCreator) CreateRouteArgsForCall(i int) (string, string, int, bool, models.DomainFields, models.SpaceFields) {
	fake.createRouteMutex.RLock()
	defer fake.createRouteMutex.RUnlock()
	return fake.createRouteArgsForCall[i].hostName, fake.createRouteArgsForCall[i].path, fake.createRouteArgsForCall[i].port, fake.createRouteArgsForCall[i].randomPort, fake.createRouteArgsForCall[i].domain, fake.createRouteArgsForCall[i].space
}

func (fake *OldFakeRouteCreator) CreateRouteReturns(result1 models.Route, result2 error) {
	fake.CreateRouteStub = nil
	fake.createRouteReturns = struct {
		result1 models.Route
		result2 error
	}{result1, result2}
}

func (cmd *OldFakeRouteCreator) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{Name: "create-route"}
}

func (cmd *OldFakeRouteCreator) SetDependency(_ commandregistry.Dependency, _ bool) commandregistry.Command {
	return cmd
}

func (cmd *OldFakeRouteCreator) Requirements(_ requirements.Factory, _ flags.FlagContext) ([]requirements.Requirement, error) {
	return []requirements.Requirement{}, nil
}

func (cmd *OldFakeRouteCreator) Execute(_ flags.FlagContext) error {
	return nil
}

var _ route.Creator = new(OldFakeRouteCreator)
