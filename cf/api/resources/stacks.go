package resources

import "github.com/cloudfoundry/cli/cf/models"

type PaginatedStackResources struct {
	Resources []StackResource
}

type StackResource struct {
	Resource
	Entity StackEntity
}

type StackEntity struct {
	Name        string
	Description string
}

func (resource StackResource) ToFields() *models.Stack {
	return &models.Stack{
		Guid:        resource.Metadata.Guid,
		Name:        resource.Entity.Name,
		Description: resource.Entity.Description,
	}
}
