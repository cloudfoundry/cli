package resources

import "cf/models"

type UserResource struct {
	Resource
	Entity UserEntity
}

type UserEntity struct {
	Name  string
	Admin bool
}

type UAAUserResources struct {
	Resources []struct {
		Id       string
		Username string
	}
}

func (resource UserResource) ToFields() models.UserFields {
	return models.UserFields{
		Guid:    resource.Metadata.Guid,
		IsAdmin: resource.Entity.Admin,
	}
}
