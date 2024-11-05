package ccv3

import (
	"net/http"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

func (requester RealRequester) paginate(request *cloudcontroller.Request, obj interface{}, appendToExternalList func(interface{}) error, specificPage bool) (IncludedResources, Warnings, error) {
	fullWarningsList := Warnings{}
	var includes IncludedResources

	for {
		wrapper, warnings, err := requester.wrapFirstPage(request, obj, appendToExternalList)
		fullWarningsList = append(fullWarningsList, warnings...)
		if err != nil {
			return IncludedResources{}, fullWarningsList, err
		}

		includes.Merge(wrapper.IncludedResources)

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
	wrapper, warnings, err := requester.wrapFirstPage(request, obj, appendToExternalList)
	if err != nil {
		return IncludedResources{}, warnings, err
	}

	if wrapper.Pagination.Next.HREF == "" {
		return wrapper.IncludedResources, warnings, nil
	}

	newQuery := url.Values{}
	for name, value := range request.URL.Query() {
		if name != "" && name != string(PerPage) {
			newQuery.Add(name, strings.Join(value, ","))
		}
	}

	newQuery.Add(string(PerPage), MaxPerPage)
	request.URL.RawQuery = newQuery.Encode()
	return requester.paginate(request, obj, appendToExternalList, false)
}
