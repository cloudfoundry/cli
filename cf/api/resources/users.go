package resources

import "github.com/cloudfoundry/cli/cf/models"

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

type UAAUserResourceEmail struct {
	Value string `json:"value"`
}

type UAAUserResourceName struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}

type UAAUserResource struct {
	Username string                 `json:"userName"`
	Emails   []UAAUserResourceEmail `json:"emails"`
	Password string                 `json:"password"`
	Name     UAAUserResourceName    `json:"name"`
}

func NewUAAUserResource(username, password string) UAAUserResource {
	return UAAUserResource{
		Username: username,
		Emails:   []UAAUserResourceEmail{{Value: username}},
		Password: password,
		Name: UAAUserResourceName{
			GivenName:  username,
			FamilyName: username,
		},
	}
}

type UAAUserFields struct {
	Id string
}
