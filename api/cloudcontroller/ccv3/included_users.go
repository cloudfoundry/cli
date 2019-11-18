package ccv3

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

// IncludedUsers represent a set of users included on an API response.
type IncludedUsers map[constant.IncludedType]User
