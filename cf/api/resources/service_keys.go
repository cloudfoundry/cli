package resources

import "github.com/cloudfoundry/cli/cf/models"

type ServiceKeyResource struct {
	Resource
	Entity ServiceKeyEntity
}

type ServiceKeyEntity struct {
	Name                string                 `json:"name"`
	ServiceInstanceGUID string                 `json:"service_instance_guid"`
	ServiceInstanceUrl  string                 `json:"service_instance_url"`
	Credentials         map[string]interface{} `json:"credentials"`
}

func (resource ServiceKeyResource) ToFields() models.ServiceKeyFields {
	return models.ServiceKeyFields{
		Name: resource.Entity.Name,
		Url:  resource.Metadata.Url,
		GUID: resource.Metadata.GUID,
	}
}

func (resource ServiceKeyResource) ToModel() models.ServiceKey {
	return models.ServiceKey{
		Fields: models.ServiceKeyFields{
			Name: resource.Entity.Name,
			GUID: resource.Metadata.GUID,
			Url:  resource.Metadata.Url,

			ServiceInstanceGUID: resource.Entity.ServiceInstanceGUID,
			ServiceInstanceUrl:  resource.Entity.ServiceInstanceUrl,
		},
		Credentials: resource.Entity.Credentials,
	}
}
