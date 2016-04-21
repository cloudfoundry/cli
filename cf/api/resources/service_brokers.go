package resources

import "github.com/cloudfoundry/cli/cf/models"

type ServiceBrokerResource struct {
	Resource
	Entity ServiceBrokerEntity
}

type ServiceBrokerEntity struct {
	GUID     string
	Name     string
	Password string `json:"auth_password"`
	Username string `json:"auth_username"`
	Url      string `json:"broker_url"`
}

func (resource ServiceBrokerResource) ToFields() (fields models.ServiceBroker) {
	fields.Name = resource.Entity.Name
	fields.GUID = resource.Metadata.GUID
	fields.Url = resource.Entity.Url
	fields.Username = resource.Entity.Username
	fields.Password = resource.Entity.Password
	return
}
