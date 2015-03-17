package resources

import "github.com/cloudfoundry/cli/cf/models"

type ServiceKeyResource struct {
	Resource
	Entity ServiceKeyEntity
}

type ServiceKeyEntity struct {
	Name                string                 `json:"name"`
	ServiceInstanceGuid string                 `json:"service_instance_guid"`
	ServiceInstanceUrl  string                 `json:"service_instance_url"`
	Credentials         map[string]interface{} `json:"credentials"`
}

func (resource ServiceKeyResource) ToFields() models.ServiceKeyFields {
	return models.ServiceKeyFields{
		Name: resource.Entity.Name,
		Url:  resource.Metadata.Url,
		Guid: resource.Metadata.Guid,
	}
}

func (resource ServiceKeyResource) ToModel() models.ServiceKey {
	return models.ServiceKey{
		Fields: models.ServiceKeyFields{
			Name: resource.Entity.Name,
			Guid: resource.Metadata.Guid,
			Url:  resource.Metadata.Url,

			ServiceInstanceGuid: resource.Entity.ServiceInstanceGuid,
			ServiceInstanceUrl:  resource.Entity.ServiceInstanceUrl,
		},
		Credentials: resource.Entity.Credentials,
	}
}
