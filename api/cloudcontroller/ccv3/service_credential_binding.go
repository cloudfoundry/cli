package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

// GetServiceCredentialBindings queries the CC API with the specified query
// and returns a slice of ServiceCredentialBindings. Additionally if Apps are
// included in the API response (by having `include=app` in the query) then the
// App names will be added into each ServiceCredentialBinding for app bindings
func (client Client) GetServiceCredentialBindings(query ...Query) ([]resources.ServiceCredentialBinding, Warnings, error) {
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
		appNameLookup := make(map[string]string)
		for _, app := range included.Apps {
			appNameLookup[app.GUID] = app.Name
		}

		for i := range result {
			result[i].AppName = appNameLookup[result[i].AppGUID]
		}
	}

	return result, warnings, err
}
