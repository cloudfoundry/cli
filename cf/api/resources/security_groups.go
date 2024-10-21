package resources

import "code.cloudfoundry.org/cli/v7/cf/models"

type PaginatedSecurityGroupResources struct {
	Resources []SecurityGroupResource
}

type SecurityGroupResource struct {
	Resource
	Entity SecurityGroup
}

type SecurityGroup struct {
	models.SecurityGroupFields
	Spaces []SpaceResource
}

func (resource SecurityGroupResource) ToFields() (fields models.SecurityGroupFields) {
	fields.Name = resource.Entity.Name
	fields.Rules = resource.Entity.Rules
	fields.SpaceURL = resource.Entity.SpaceURL
	fields.GUID = resource.Metadata.GUID

	return
}

func (resource SecurityGroupResource) ToModel() (asg models.SecurityGroup) {
	asg.SecurityGroupFields = resource.ToFields()
	return
}
