package ccv3

import (
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/lookuptable"
)

// GetServiceOffering lists service offering with optional filters.
func (client *Client) GetServiceOfferings(query ...Query) ([]resources.ServiceOffering, Warnings, error) {
	var result []resources.ServiceOffering

	query = append(query, Query{Key: FieldsServiceBroker, Values: []string{"name", "guid"}})

	included, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetServiceOfferingsRequest,
		Query:        query,
		ResponseBody: resources.ServiceOffering{},
		AppendToList: func(item interface{}) error {
			result = append(result, item.(resources.ServiceOffering))
			return nil
		},
	})

	brokerNameLookup := lookuptable.NameFromGUID(included.ServiceBrokers)

	for i, _ := range result {
		result[i].ServiceBrokerName = brokerNameLookup[result[i].ServiceBrokerGUID]
	}

	return result, warnings, err
}

func (client *Client) GetServiceOfferingByGUID(guid string) (resources.ServiceOffering, Warnings, error) {
	if guid == "" {
		return resources.ServiceOffering{}, nil, ccerror.ServiceOfferingNotFoundError{}
	}

	var result resources.ServiceOffering

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetServiceOfferingRequest,
		URIParams:    internal.Params{"service_offering_guid": guid},
		ResponseBody: &result,
	})

	return result, warnings, err
}

func (client *Client) GetServiceOfferingByNameAndBroker(serviceOfferingName, serviceBrokerName string) (resources.ServiceOffering, Warnings, error) {
	query := []Query{
		{Key: NameFilter, Values: []string{serviceOfferingName}},
		{Key: PerPage, Values: []string{"2"}},
		{Key: Page, Values: []string{"1"}},
	}
	if serviceBrokerName != "" {
		query = append(query, Query{Key: ServiceBrokerNamesFilter, Values: []string{serviceBrokerName}})
	}

	offerings, warnings, err := client.GetServiceOfferings(query...)
	if err != nil {
		return resources.ServiceOffering{}, warnings, err
	}

	switch len(offerings) {
	case 0:
		return resources.ServiceOffering{}, warnings, ccerror.ServiceOfferingNotFoundError{
			ServiceOfferingName: serviceOfferingName,
			ServiceBrokerName:   serviceBrokerName,
		}
	case 1:
		return offerings[0], warnings, nil
	default:
		return resources.ServiceOffering{}, warnings, ccerror.ServiceOfferingNameAmbiguityError{
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

func extractServiceBrokerNames(offerings []resources.ServiceOffering) (result []string) {
	for _, o := range offerings {
		result = append(result, o.ServiceBrokerName)
	}
	return
}
