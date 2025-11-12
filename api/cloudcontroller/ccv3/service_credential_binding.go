package ccv3

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/lookuptable"
)

func (client *Client) CreateServiceCredentialBinding(binding resources.ServiceCredentialBinding) (JobURL, Warnings, error) {
	return client.MakeRequest(RequestParams{
		RequestName: internal.PostServiceCredentialBindingRequest,
		RequestBody: binding,
	})
}

// GetServiceCredentialBindings queries the CC API with the specified query
// and returns a slice of ServiceCredentialBindings. Additionally if Apps are
// included in the API response (by having `include=app` in the query) then the
// App names will be added into each ServiceCredentialBinding for app bindings
func (client *Client) GetServiceCredentialBindings(query ...Query) ([]resources.ServiceCredentialBinding, Warnings, error) {
	var result []resources.ServiceCredentialBinding

	included, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetServiceCredentialBindingsRequest,
		Query:        query,
		ResponseBody: resources.ServiceCredentialBinding{},
		AppendToList: func(item interface{}) error {
			result = append(result, item.(resources.ServiceCredentialBinding))
			return nil
		},
	})

	if len(included.Apps) > 0 {
		appLookup := lookuptable.AppFromGUID(included.Apps)

		for i := range result {
			result[i].AppName = appLookup[result[i].AppGUID].Name
			result[i].AppSpaceGUID = appLookup[result[i].AppGUID].SpaceGUID
		}
	}

	return result, warnings, err
}

func (client *Client) DeleteServiceCredentialBinding(guid string) (JobURL, Warnings, error) {
	return client.MakeRequest(RequestParams{
		RequestName: internal.DeleteServiceCredentialBindingRequest,
		URIParams:   internal.Params{"service_credential_binding_guid": guid},
	})
}

func (client *Client) GetServiceCredentialBindingDetails(guid string) (details resources.ServiceCredentialBindingDetails, warnings Warnings, err error) {
	_, warnings, err = client.MakeRequest(RequestParams{
		RequestName:  internal.GetServiceCredentialBindingDetailsRequest,
		URIParams:    internal.Params{"service_credential_binding_guid": guid},
		ResponseBody: &details,
	})

	return
}
