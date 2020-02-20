package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
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

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.PostSpaceRequest,
		RequestBody:  space,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) DeleteSpace(spaceGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.makeRequest(requestParams{
		RequestName: internal.DeleteSpaceRequest,
		URIParams:   internal.Params{"space_guid": spaceGUID},
	})

	return jobURL, warnings, err
}

// GetSpaces lists spaces with optional filters.
func (client *Client) GetSpaces(query ...Query) ([]Space, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpacesRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullSpacesList []Space
	warnings, err := client.paginate(request, Space{}, func(item interface{}) error {
		if space, ok := item.(Space); ok {
			fullSpacesList = append(fullSpacesList, space)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Space{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullSpacesList, warnings, err
}

func (client *Client) UpdateSpace(space Space) (Space, Warnings, error) {
	spaceGUID := space.GUID
	space.GUID = ""
	space.Relationships = nil

	var responseBody Space

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.PatchSpaceRequest,
		URIParams:    map[string]string{"space_guid": spaceGUID},
		RequestBody:  space,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
