package resources

import "github.com/cloudfoundry/cli/cf/models"

type PaginatedApplicationSecurityGroupResources struct {
	Resources []ApplicationSecurityGroupResource
}

type ApplicationSecurityGroupResource struct {
	Resource
	Entity ApplicationSecurityGroup
}

// represents a fully instantiated model returned by the CC (e.g.: with its attributes and the fields for its child objects)
type ApplicationSecurityGroup struct {
	models.ApplicationSecurityGroupFields
	Spaces []SpaceResource
}

func (resource ApplicationSecurityGroupResource) ToFields() (fields models.ApplicationSecurityGroupFields) {
	fields.Name = resource.Entity.Name
	fields.Rules = resource.Entity.Rules
	fields.Guid = resource.Metadata.Guid

	return
}

func (resource ApplicationSecurityGroupResource) ToModel() (asg models.ApplicationSecurityGroup) {
	asg.ApplicationSecurityGroupFields = resource.ToFields()

	spaces := []models.Space{}
	for _, s := range resource.Entity.Spaces {
		spaces = append(spaces, s.ToModel())
	}
	asg.Spaces = spaces

	return
}
