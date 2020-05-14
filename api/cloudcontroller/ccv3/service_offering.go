package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/cloudcontroller/jsonry"
	"code.cloudfoundry.org/cli/resources"
)

// ServiceOffering represents a Cloud Controller V3 Service Offering.
type ServiceOffering struct {
	// GUID is a unique service offering identifier.
	GUID string
	// Name is the name of the service offering.
	Name string
	// ServiceBrokerName is the name of the service broker
	ServiceBrokerName string
	// ServiceBrokerGUID is the guid of the service broker
	ServiceBrokerGUID string `jsonry:"relationships.service_broker.data.guid"`
	// Description of the service offering
	Description string

	Metadata *resources.Metadata
}

func (so *ServiceOffering) UnmarshalJSON(data []byte) error {
	return jsonry.Unmarshal(data, so)
}

// GetServiceOffering lists service offering with optional filters.
func (client *Client) GetServiceOfferings(query ...Query) ([]ServiceOffering, Warnings, error) {
	var resources []ServiceOffering

	query = append(query, Query{Key: FieldsServiceBroker, Values: []string{"name", "guid"}})

	included, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetServiceOfferingsRequest,
		Query:        query,
		ResponseBody: ServiceOffering{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(ServiceOffering))
			return nil
		},
	})

	brokerNameLookup := make(map[string]string)
	for _, b := range included.ServiceBrokers {
		brokerNameLookup[b.GUID] = b.Name
	}

	for i, _ := range resources {
		resources[i].ServiceBrokerName = brokerNameLookup[resources[i].ServiceBrokerGUID]
	}

	return resources, warnings, err
}

func (client *Client) GetServiceOfferingByNameAndBroker(serviceOfferingName, serviceBrokerName string) (ServiceOffering, Warnings, error) {
	query := []Query{{Key: NameFilter, Values: []string{serviceOfferingName}}}
	if serviceBrokerName != "" {
		query = append(query, Query{Key: ServiceBrokerNamesFilter, Values: []string{serviceBrokerName}})
	}

	offerings, warnings, err := client.GetServiceOfferings(query...)
	if err != nil {
		return ServiceOffering{}, warnings, err
	}

	switch len(offerings) {
	case 0:
		return ServiceOffering{}, warnings, ccerror.ServiceOfferingNotFoundError{
			ServiceOfferingName: serviceOfferingName,
			ServiceBrokerName:   serviceBrokerName,
		}
	case 1:
		return offerings[0], warnings, nil
	default:
		return ServiceOffering{}, warnings, ccerror.ServiceOfferingNameAmbiguityError{
			ServiceOfferingName: serviceOfferingName,
			ServiceBrokerNames:  extractServiceBrokerNames(offerings),
		}
	}
}

func (client *Client) PurgeServiceOffering(serviceOfferingGUID string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteServiceOfferingRequest,
		URIParams:   internal.Params{"service_offering_guid": serviceOfferingGUID},
		Query:       []Query{{Key: Purge, Values: []string{"true"}}},
	})
	return warnings, err
}

func extractServiceBrokerNames(offerings []ServiceOffering) (result []string) {
	for _, o := range offerings {
		result = append(result, o.ServiceBrokerName)
	}
	return
}
