package commands

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/simonleung8/flags"
)

type FakeRouteCreator struct {
	CreateRouteHostname     string
	CreateRouteDomainFields models.DomainFields
	CreateRouteSpaceFields  models.SpaceFields
	ReservedRoute           models.Route
}

func (cmd *FakeRouteCreator) CreateRoute(hostName string, domain models.DomainFields, space models.SpaceFields) (reservedRoute models.Route, apiErr error) {
	cmd.CreateRouteHostname = hostName
	cmd.CreateRouteDomainFields = domain
	cmd.CreateRouteSpaceFields = space
	reservedRoute = cmd.ReservedRoute
	return
}

func (cmd *FakeRouteCreator) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{Name: "create-route"}
}

func (cmd *FakeRouteCreator) SetDependency(_ command_registry.Dependency, _ bool) command_registry.Command {
	return cmd
}

func (cmd *FakeRouteCreator) Requirements(_ requirements.Factory, _ flags.FlagContext) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd *FakeRouteCreator) Execute(_ flags.FlagContext) {}
