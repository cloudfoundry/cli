package ccv3

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

func (requester RealRequester) paginate(request *cloudcontroller.Request, obj interface{}, appendToExternalList func(interface{}) error) (IncludedResources, Warnings, error) {
	fullWarningsList := Warnings{}
	var includes IncludedResources

	for {
		wrapper, warnings, err := requester.wrapFirstPage(request, obj, appendToExternalList)
		fullWarningsList = append(fullWarningsList, warnings...)
		if err != nil {
			return IncludedResources{}, fullWarningsList, err
		}

		includes.Users = append(includes.Users, wrapper.IncludedResources.Users...)
		includes.Organizations = append(includes.Organizations, wrapper.IncludedResources.Organizations...)
		includes.Spaces = append(includes.Spaces, wrapper.IncludedResources.Spaces...)
		includes.ServiceOfferings = append(includes.ServiceOfferings, wrapper.IncludedResources.ServiceOfferings...)
		includes.ServiceBrokers = append(includes.ServiceBrokers, wrapper.IncludedResources.ServiceBrokers...)

		if wrapper.NextPage() == "" {
			break
		}

		request, err = requester.newHTTPRequest(requestOptions{
			URL:    wrapper.NextPage(),
			Method: http.MethodGet,
		})
		if err != nil {
			return IncludedResources{}, fullWarningsList, err
		}
	}

	return includes, fullWarningsList, nil
}

func (requester RealRequester) wrapFirstPage(request *cloudcontroller.Request, obj interface{}, appendToExternalList func(interface{}) error) (*PaginatedResources, Warnings, error) {
	warnings := Warnings{}
	wrapper := NewPaginatedResources(obj)
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &wrapper,
	}

	err := requester.connection.Make(request, &response)
	warnings = append(warnings, response.Warnings...)
	if err != nil {
		return nil, warnings, err
	}

	list, err := wrapper.Resources()
	if err != nil {
		return nil, warnings, err
	}

	for _, item := range list {
		err = appendToExternalList(item)
		if err != nil {
			return nil, warnings, err
		}
	}

	return wrapper, warnings, nil
}
