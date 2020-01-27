package ccv3

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"encoding/json"
)

type SpaceQuota struct {
	// GUID is the unique ID of the space quota.
	GUID string `json:"guid,omitempty"`
	// Name is the name of the space quota
	Name string `json:"name"`
	// Apps contain the various limits that are associated with applications
	Apps AppLimit `json:"apps"`
	// Services contain the various limits that are associated with services
	Services ServiceLimit `json:"services"`
	// Routes contain the various limits that are associated with routes
	Routes RouteLimit `json:"routes"`
	// OrgGUID is the unique ID of the owning organization
	OrgGUID    string
	// SpaceGUIDs are the list of unique ID's of the associated spaces
	SpaceGUIDs []string
}

func (sq SpaceQuota) MarshalJSON() ([]byte, error) {
	type Data struct {
		GUID string `json:"guid,omitempty"`
	}

	type SpacesData struct {
		Data []Data `json:"data,omitempty"`
	}

	type OrganizationData struct {
		Data Data `json:"data,omitempty"`
	}

	type relationships struct {
		Organization OrganizationData `json:"organization,omitempty"`
		Spaces       SpacesData       `json:"spaces,omitempty"`
	}

	type ccSpaceQuota struct {
		GUID          string         `json:"guid,omitempty"`
		Name          string         `json:"name"`
		Relationships *relationships `json:"relationships"`
	}

	ccSQ := ccSpaceQuota{
		Name: sq.Name,
		GUID: sq.GUID,
		Relationships: &relationships{
			Organization: OrganizationData{
				Data: Data{
					GUID: sq.OrgGUID,
				},
			},
		},
	}

	for _, spaceGUID := range sq.SpaceGUIDs {
		ccSQ.Relationships.Spaces.Data = append(ccSQ.Relationships.Spaces.Data, Data{GUID: spaceGUID})
	}

	return json.Marshal(ccSQ)
}

func (sq *SpaceQuota) UnmarshalJSON(data []byte) error {
	var ccSpaceQuotaStruct struct {
		GUID          string         `json:"guid,omitempty"`
		Name          string         `json:"name"`
		Relationships struct {
			Organization struct {
				Data struct {
					GUID string `json:"guid,omitempty"`
				} `json:"data,omitempty"`
			} `json:"organization,omitempty"`
			Spaces struct {
				Data []struct {
					GUID string `json:"guid,omitempty"`
				} `json:"data,omitempty"`
			}`json:"spaces,omitempty"`
		} `json:"relationships,omitempty"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccSpaceQuotaStruct)
	if err != nil {
		return err
	}

	sq.GUID = ccSpaceQuotaStruct.GUID
	sq.Name = ccSpaceQuotaStruct.Name
	sq.OrgGUID = ccSpaceQuotaStruct.Relationships.Organization.Data.GUID

	for _, spaceData := range ccSpaceQuotaStruct.Relationships.Spaces.Data {
		sq.SpaceGUIDs = append(sq.SpaceGUIDs, string(spaceData.GUID))
	}

	return nil
}

func (client Client) CreateSpaceQuota(spaceQuota SpaceQuota) (SpaceQuota, Warnings, error) {
	spaceQuotaBytes, err := json.Marshal(spaceQuota)

	if err != nil {
		return SpaceQuota{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostSpaceQuotaRequest,
		Body:        bytes.NewReader(spaceQuotaBytes),
	})

	if err != nil {
		return SpaceQuota{}, nil, err
	}

	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &spaceQuota,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return SpaceQuota{}, response.Warnings, err
	}

	return spaceQuota, response.Warnings, err
}
