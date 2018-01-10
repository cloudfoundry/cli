package ccv3

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

// Relationships is a map of RelationshipTypes to Relationship.
type Relationships map[constant.RelationshipType]Relationship
