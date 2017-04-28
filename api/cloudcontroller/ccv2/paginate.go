package ccv2

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

func (client Client) paginate(request *cloudcontroller.Request, obj interface{}, appendToExternalList func(interface{}) error) (Warnings, error) {
	fullWarningsList := Warnings{}

	for {
		wrapper := NewPaginatedResources(obj)
		response := cloudcontroller.Response{
			Result: &wrapper,
		}

		err := client.connection.Make(request, &response)
		fullWarningsList = append(fullWarningsList, response.Warnings...)
		if err != nil {
			return fullWarningsList, err
		}

		list, err := wrapper.Resources()
		if err != nil {
			return fullWarningsList, err
		}

		for _, item := range list {
			err = appendToExternalList(item)
			if err != nil {
				return fullWarningsList, err
			}
		}

		if wrapper.NextURL == "" {
			break
		}

		request, err = client.newHTTPRequest(requestOptions{
			URI:    wrapper.NextURL,
			Method: http.MethodGet,
		})
		if err != nil {
			return fullWarningsList, err
		}
	}

	return fullWarningsList, nil
}
