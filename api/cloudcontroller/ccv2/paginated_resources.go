package ccv2

import (
	"encoding/json"
	"reflect"
)

// NewPaginatedResources returns a new PaginatedResources struct with the
// given resource type.
func NewPaginatedResources(exampleResource interface{}) *PaginatedResources {
	return &PaginatedResources{
		resourceType: reflect.TypeOf(exampleResource),
	}
}

// PaginatedResources represents a page of resources returned by the Cloud
// Controller.
type PaginatedResources struct {
	NextURL        string          `json:"next_url"`
	ResourcesBytes json.RawMessage `json:"resources"`
	resourceType   reflect.Type
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
