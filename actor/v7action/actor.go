// Package v7action contains the business logic for the commands/v7 package
package v7action

import (
	"code.cloudfoundry.org/cli/v8/api/uaa/constant"
	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/clock"
)

// SortOrder is used for sorting.
type SortOrder string

const (
	Ascending  SortOrder = "Ascending"
	Descending SortOrder = "Descending"
)

// Warnings is a list of warnings returned back from the cloud controller
type Warnings []string

type AuthActor interface {
	Authenticate(credentials map[string]string, origin string, grantType constant.GrantType) error
	GetLoginPrompts() (map[string]coreconfig.AuthPrompt, error)
	GetCurrentUser() (configv3.User, error)
}

// Actor represents a V7 actor.
type Actor struct {
	CloudControllerClient CloudControllerClient
	Config                Config
	SharedActor           SharedActor
	UAAClient             UAAClient
	RoutingClient         RoutingClient
	Clock                 clock.Clock

	AuthActor
}

// NewActor returns a new V7 actor.
func NewActor(
	client CloudControllerClient,
	config Config,
	sharedActor SharedActor,
	uaaClient UAAClient,
	routingClient RoutingClient,
	clk clock.Clock,
) *Actor {
	authActor := NewDefaultAuthActor(config, uaaClient)
	if config != nil && config.IsCFOnK8s() {
		authActor = NewKubernetesAuthActor(config, NewDefaultKubernetesConfigGetter(), client)
	}

	return &Actor{
		CloudControllerClient: client,
		Config:                config,
		SharedActor:           sharedActor,
		UAAClient:             uaaClient,
		RoutingClient:         routingClient,
		Clock:                 clk,
		AuthActor:             authActor,
	}
}
