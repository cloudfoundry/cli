package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client Client) GetServiceCredentialBindings(query ...Query) ([]resources.ServiceCredentialBinding, IncludedResources, Warnings, error) {
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

	return result, included, warnings, err
}
