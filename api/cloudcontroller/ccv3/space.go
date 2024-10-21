package ccv3

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v9/resources"
)

func (client *Client) CreateSpace(space resources.Space) (resources.Space, Warnings, error) {
	var responseBody resources.Space

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
func (client *Client) GetSpaces(query ...Query) ([]resources.Space, IncludedResources, Warnings, error) {
	var returnedResources []resources.Space

	includedResources, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetSpacesRequest,
		Query:        query,
		ResponseBody: resources.Space{},
		AppendToList: func(item interface{}) error {
			returnedResources = append(returnedResources, item.(resources.Space))
			return nil
		},
	})

	return returnedResources, includedResources, warnings, err
}

func (client *Client) UpdateSpace(space resources.Space) (resources.Space, Warnings, error) {
	spaceGUID := space.GUID
	space.GUID = ""
	space.Relationships = nil

	var responseBody resources.Space

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchSpaceRequest,
		URIParams:    internal.Params{"space_guid": spaceGUID},
		RequestBody:  space,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
