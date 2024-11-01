package ccv3

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

const MaxResultsPerPage int = 5000

func (requester RealRequester) paginate(request *cloudcontroller.Request, obj interface{}, appendToExternalList func(interface{}) error, specificPage bool) (IncludedResources, Warnings, error) {
	fullWarningsList := Warnings{}
	var includes IncludedResources

	for {
		wrapper, warnings, err := requester.wrapFirstPage(request, obj, appendToExternalList)
		fullWarningsList = append(fullWarningsList, warnings...)
		if err != nil {
			return IncludedResources{}, fullWarningsList, err
		}

		includes.Apps = append(includes.Apps, wrapper.IncludedResources.Apps...)
		includes.Users = append(includes.Users, wrapper.IncludedResources.Users...)
		includes.Organizations = append(includes.Organizations, wrapper.IncludedResources.Organizations...)
		includes.Spaces = append(includes.Spaces, wrapper.IncludedResources.Spaces...)
		includes.ServiceBrokers = append(includes.ServiceBrokers, wrapper.IncludedResources.ServiceBrokers...)
		includes.ServiceInstances = append(includes.ServiceInstances, wrapper.IncludedResources.ServiceInstances...)
		includes.ServiceOfferings = append(includes.ServiceOfferings, wrapper.IncludedResources.ServiceOfferings...)
		includes.ServicePlans = append(includes.ServicePlans, wrapper.IncludedResources.ServicePlans...)

		if specificPage || wrapper.NextPage() == "" {
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

func (requester RealRequester) bulkRetrieval(request *cloudcontroller.Request, obj interface{}, appendToExternalList func(interface{}) error) (IncludedResources, Warnings, error) {
	wrapper, warnings, err := requester.wrapFirstPage(request, obj, func(_ interface{}) error { return nil })
	if err != nil {
		return IncludedResources{}, warnings, err
	}

	newQuery := url.Values{}
	for name, value := range request.URL.Query() {
		if name != "" && name != string(PerPage) {
			newQuery.Add(name, strings.Join(value, ","))
		}
	}
	resultsPerPage := strconv.Itoa(wrapper.Pagination.TotalResults)
	if wrapper.Pagination.TotalResults > MaxResultsPerPage {
		resultsPerPage = strconv.Itoa(MaxResultsPerPage)
	}
	newQuery.Add(string(PerPage), resultsPerPage)
	request.URL.RawQuery = newQuery.Encode()
	return requester.paginate(request, obj, appendToExternalList, false)
}
