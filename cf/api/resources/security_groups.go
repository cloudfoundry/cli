package resources

import "github.com/cloudfoundry/cli/cf/models"

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
	fields.SpaceUrl = resource.Entity.SpaceUrl
	fields.Guid = resource.Metadata.Guid

	return
}

func (resource SecurityGroupResource) ToModel() (asg models.SecurityGroup) {
	asg.SecurityGroupFields = resource.ToFields()
	return
}
