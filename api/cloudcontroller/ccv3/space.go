package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Space represents a Cloud Controller V3 Space.
type Space struct {
	// GUID is a unique space identifier.
	GUID string `json:"guid,omitempty"`
	// Name is the name of the space.
	Name string `json:"name"`
	// Relationships list the relationships to the space.
	Relationships Relationships `json:"relationships,omitempty"`
	// Metadata is used for custom tagging of API resources
	Metadata *Metadata `json:"metadata,omitempty"`
}

func (client *Client) CreateSpace(space Space) (Space, Warnings, error) {
	var responseBody Space

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostSpaceRequest,
		RequestBody:  space,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) DeleteSpace(spaceGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteSpaceRequest,
		URIParams:   internal.Params{"space_guid": spaceGUID},
	})

	return jobURL, warnings, err
}

// GetSpaces lists spaces with optional filters.
func (client *Client) GetSpaces(query ...Query) ([]Space, IncludedResources, Warnings, error) {
	var resources []Space

	includedResources, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetSpacesRequest,
		Query:        query,
		ResponseBody: Space{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Space))
			return nil
		},
	})

	return resources, includedResources, warnings, err
}

func (client *Client) UpdateSpace(space Space) (Space, Warnings, error) {
	spaceGUID := space.GUID
	space.GUID = ""
	space.Relationships = nil

	var responseBody Space

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchSpaceRequest,
		URIParams:    internal.Params{"space_guid": spaceGUID},
		RequestBody:  space,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
