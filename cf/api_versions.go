package cf

import "github.com/blang/semver/v4"

var (
	UserProvidedServiceTagsMinimumAPIVersion, _ = semver.Make("2.104.0") // #158233239,#157770881

	ServiceAuthTokenMaximumAPIVersion, _ = semver.Make("2.46.0")
)
