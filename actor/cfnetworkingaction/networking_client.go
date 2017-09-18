package cfnetworkingaction

import "code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1"

//go:generate counterfeiter . NetworkingClient
type NetworkingClient interface {
	CreatePolicies(policies []cfnetv1.Policy) error
	ListPolicies(appNames ...string) ([]cfnetv1.Policy, error)
	RemovePolicies(policies []cfnetv1.Policy) error
}
