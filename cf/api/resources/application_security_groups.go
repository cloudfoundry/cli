package resources

import "github.com/cloudfoundry/cli/cf/models"

type PaginatedApplicationSecurityGroupResources struct {
	Resources []ApplicationSecurityGroupResource
}

type ApplicationSecurityGroupResource struct {
	Resource
	Entity models.ApplicationSecurityGroupFields
}

func (resource ApplicationSecurityGroupResource) ToFields() models.ApplicationSecurityGroupFields {
	appSecurityGroup := resource.Entity
	appSecurityGroup.Guid = resource.Metadata.Guid

	return appSecurityGroup
}
