package resources

import "code.cloudfoundry.org/cli/v7/cf/models"

type ServicePlanVisibilityResource struct {
	Resource
	Entity models.ServicePlanVisibilityFields
}

func (resource ServicePlanVisibilityResource) ToFields() (fields models.ServicePlanVisibilityFields) {
	fields.GUID = resource.Metadata.GUID
	fields.ServicePlanGUID = resource.Entity.ServicePlanGUID
	fields.OrganizationGUID = resource.Entity.OrganizationGUID
	return
}
