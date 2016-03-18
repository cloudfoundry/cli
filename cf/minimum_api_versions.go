package cf

import "github.com/blang/semver"

var (
	TcpRoutingMinimumApiVersion, _                      = semver.Make("2.51.0")
	MultipleAppPortsMinimumApiVersion, _                = semver.Make("2.51.0")
	UpdateServicePlanMinimumApiVersion, _               = semver.Make("2.16.0")
	SetRolesByUsernameMinimumApiVersion, _              = semver.Make("2.37.0")
	ListUsersInOrgOrSpaceWithoutUAAMinimumApiVersion, _ = semver.Make("2.21.0")
	RoutePathMinimumApiVersion, _                       = semver.Make("2.36.0")
	OrgAppInstanceLimitMinimumApiVersion, _             = semver.Make("2.33.0")
	SpaceAppInstanceLimitMinimumApiVersion, _           = semver.Make("2.40.0")
)
