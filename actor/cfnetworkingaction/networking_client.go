package cfnetworkingaction

import "code.cloudfoundry.org/cli/v9/api/cfnetworking/cfnetv1"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . NetworkingClient

type NetworkingClient interface {
	CreatePolicies(policies []cfnetv1.Policy) error
	ListPolicies(appGUIDs ...string) ([]cfnetv1.Policy, error)
	RemovePolicies(policies []cfnetv1.Policy) error
}
