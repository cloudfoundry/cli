package cfnetworkingaction

import "code.cloudfoundry.org/cli/api/cfnetworking/cfnetv1"

//go:generate counterfeiter . NetworkingClient
type NetworkingClient interface {
	CreatePolicies(policies []cfnetv1.Policy) error
}
