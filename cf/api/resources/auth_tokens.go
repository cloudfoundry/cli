package resources

import "github.com/cloudfoundry/cli/cf/models"

type PaginatedAuthTokenResources struct {
	Resources []AuthTokenResource
}

type AuthTokenResource struct {
	Resource
	Entity AuthTokenEntity
}

type AuthTokenEntity struct {
	Label    string
	Provider string
}

func (resource AuthTokenResource) ToFields() (authToken models.ServiceAuthTokenFields) {
	authToken.Guid = resource.Metadata.Guid
	authToken.Label = resource.Entity.Label
	authToken.Provider = resource.Entity.Provider
	return
}
