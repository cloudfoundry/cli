package ccv2

import (
	"encoding/json"
	"net/http"
	"reflect"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

// PaginatedResources represents a page of resources returned by the Cloud
// Controller.
type PaginatedResources struct {
	NextURL        string          `json:"next_url"`
	ResourcesBytes json.RawMessage `json:"resources"`
	resourceType   reflect.Type
}

// NewPaginatedResources returns a new PaginatedResources struct with the
// given resource type.
func NewPaginatedResources(exampleResource interface{}) *PaginatedResources {
	return &PaginatedResources{
		resourceType: reflect.TypeOf(exampleResource),
	}
}

// Resources unmarshals JSON representing a page of resources and returns a
// slice of the given resource type.
func (pr PaginatedResources) Resources() ([]interface{}, error) {
	slicePtr := reflect.New(reflect.SliceOf(pr.resourceType))
	err := json.Unmarshal([]byte(pr.ResourcesBytes), slicePtr.Interface())
	slice := reflect.Indirect(slicePtr)

	contents := make([]interface{}, 0, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		contents = append(contents, slice.Index(i).Interface())
	}
	return contents, err
}

func (client Client) paginate(request *cloudcontroller.Request, obj interface{}, appendToExternalList func(interface{}) error) (Warnings, error) {
	fullWarningsList := Warnings{}

	for {
		wrapper := NewPaginatedResources(obj)
		response := cloudcontroller.Response{
			DecodeJSONResponseInto: &wrapper,
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
