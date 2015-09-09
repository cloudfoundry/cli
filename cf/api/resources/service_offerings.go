package resources

import (
	"encoding/json"
	"strconv"

	"github.com/cloudfoundry/cli/cf/models"
)

type PaginatedServiceOfferingResources struct {
	Resources []ServiceOfferingResource
}

type ServiceOfferingResource struct {
	Resource
	Entity ServiceOfferingEntity
}

type ServiceOfferingEntity struct {
	Label        string                `json:"label"`
	Version      string                `json:"version"`
	Description  string                `json:"description"`
	Provider     string                `json:"provider"`
	BrokerGuid   string                `json:"service_broker_guid"`
	ServicePlans []ServicePlanResource `json:"service_plans"`
	Extra        ServiceOfferingExtra
}

type ServiceOfferingExtra struct {
	DocumentationURL string `json:"documentationUrl"`
}

func (resource ServiceOfferingResource) ToFields() (fields models.ServiceOfferingFields) {
	fields.Label = resource.Entity.Label
	fields.Version = resource.Entity.Version
	fields.Provider = resource.Entity.Provider
	fields.Description = resource.Entity.Description
	fields.BrokerGuid = resource.Entity.BrokerGuid
	fields.Guid = resource.Metadata.Guid
	fields.DocumentationUrl = resource.Entity.Extra.DocumentationURL

	return
}

func (resource ServiceOfferingResource) ToModel() (offering models.ServiceOffering) {
	offering.ServiceOfferingFields = resource.ToFields()
	for _, p := range resource.Entity.ServicePlans {
		servicePlan := models.ServicePlanFields{}
		servicePlan.Name = p.Entity.Name
		servicePlan.Guid = p.Metadata.Guid
		offering.Plans = append(offering.Plans, servicePlan)
	}
	return offering
}

type serviceOfferingExtra ServiceOfferingExtra

func (resource *ServiceOfferingExtra) UnmarshalJSON(rawData []byte) error {
	if string(rawData) == "null" {
		return nil
	}

	extra := serviceOfferingExtra{}

	unquoted, err := strconv.Unquote(string(rawData))
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(unquoted), &extra)
	if err != nil {
		return err
	}

	*resource = ServiceOfferingExtra(extra)

	return nil
}
