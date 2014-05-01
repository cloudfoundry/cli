package resources

import "github.com/cloudfoundry/cli/cf/models"

type ServiceBindingResource struct {
	Resource
	Entity ServiceBindingEntity
}

type ServiceBindingEntity struct {
	AppGuid string `json:"app_guid"`
}

func (resource ServiceBindingResource) ToFields() (fields models.ServiceBindingFields) {
	fields.Url = resource.Metadata.Url
	fields.Guid = resource.Metadata.Guid
	fields.AppGuid = resource.Entity.AppGuid
	return
}
