package ccv3

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
	// Pagination represents information about the paginated resource.
	Pagination struct {
		// Next represents a link to the next page.
		Next struct {
			// HREF is the HREF of the next page.
			HREF string `json:"href"`
		} `json:"next"`
	} `json:"pagination"`
	// ResourceBytes is the list of resources for the current page.
	ResourcesBytes    json.RawMessage `json:"resources"`
	resourceType      reflect.Type
	IncludedResources struct {
		UserResource []User `json:"users"`
	} `json:"included"`
}

// NextPage returns the HREF of the next page of results.
func (pr PaginatedResources) NextPage() string {
	return pr.Pagination.Next.HREF
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
