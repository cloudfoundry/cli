package cf

import "github.com/blang/semver"

var (
	UserProvidedServiceTagsMinimumAPIVersion, _         = semver.Make("2.104.0") // #158233239,#157770881
	SetRolesByUsernameMinimumAPIVersion, _              = semver.Make("2.37.0")
	RoutePathMinimumAPIVersion, _                       = semver.Make("2.36.0")
	OrgAppInstanceLimitMinimumAPIVersion, _             = semver.Make("2.33.0")
	ListUsersInOrgOrSpaceWithoutUAAMinimumAPIVersion, _ = semver.Make("2.21.0")
	UpdateServicePlanMinimumAPIVersion, _               = semver.Make("2.16.0")

	ServiceAuthTokenMaximumAPIVersion, _ = semver.Make("2.46.0")
	SpaceScopedMaximumAPIVersion, _      = semver.Make("2.47.0")
)
