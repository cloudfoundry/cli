package resources

import "github.com/cloudfoundry/cli/cf/models"

type ServiceBindingResource struct {
	Resource
	Entity ServiceBindingEntity
}

type ServiceBindingEntity struct {
	AppGUID string `json:"app_guid"`
}

func (resource ServiceBindingResource) ToFields() (fields models.ServiceBindingFields) {
	fields.URL = resource.Metadata.URL
	fields.GUID = resource.Metadata.GUID
	fields.AppGUID = resource.Entity.AppGUID
	return
}
