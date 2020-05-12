package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client Client) ApplySpaceQuota(quotaGUID string, spaceGUID string) (resources.RelationshipList, Warnings, error) {
	var responseBody resources.RelationshipList

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostSpaceQuotaRelationshipsRequest,
		URIParams:    internal.Params{"quota_guid": quotaGUID},
		RequestBody:  resources.RelationshipList{GUIDs: []string{spaceGUID}},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client Client) CreateSpaceQuota(spaceQuota resources.SpaceQuota) (resources.SpaceQuota, Warnings, error) {
	var responseBody resources.SpaceQuota

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostSpaceQuotaRequest,
		RequestBody:  spaceQuota,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client Client) DeleteSpaceQuota(spaceQuotaGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteSpaceQuotaRequest,
		URIParams:   internal.Params{"quota_guid": spaceQuotaGUID},
	})

	return jobURL, warnings, err
}

func (client Client) GetSpaceQuota(spaceQuotaGUID string) (resources.SpaceQuota, Warnings, error) {
	var responseBody resources.SpaceQuota

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetSpaceQuotaRequest,
		URIParams:    internal.Params{"quota_guid": spaceQuotaGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) GetSpaceQuotas(query ...Query) ([]resources.SpaceQuota, Warnings, error) {
	var spaceQuotas []resources.SpaceQuota

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetSpaceQuotasRequest,
		Query:        query,
		ResponseBody: resources.SpaceQuota{},
		AppendToList: func(item interface{}) error {
			spaceQuotas = append(spaceQuotas, item.(resources.SpaceQuota))
			return nil
		},
	})

	return spaceQuotas, warnings, err
}

func (client *Client) UnsetSpaceQuota(spaceQuotaGUID, spaceGUID string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteSpaceQuotaFromSpaceRequest,
		URIParams:   internal.Params{"quota_guid": spaceQuotaGUID, "space_guid": spaceGUID},
	})

	return warnings, err
}

func (client *Client) UpdateSpaceQuota(spaceQuota resources.SpaceQuota) (resources.SpaceQuota, Warnings, error) {
	spaceQuotaGUID := spaceQuota.GUID
	spaceQuota.GUID = ""

	var responseBody resources.SpaceQuota

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchSpaceQuotaRequest,
		URIParams:    internal.Params{"quota_guid": spaceQuotaGUID},
		RequestBody:  spaceQuota,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
