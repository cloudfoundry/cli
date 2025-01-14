package ccv3

import (
	ccv3internal "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client *Client) ApplyOrganizationQuota(quotaGuid, orgGuid string) (resources.RelationshipList, Warnings, error) {
	var responseBody resources.RelationshipList

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.PostOrganizationQuotaApplyRequest,
		URIParams:    internal.Params{"quota_guid": quotaGuid},
		RequestBody:  resources.RelationshipList{GUIDs: []string{orgGuid}},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) CreateOrganizationQuota(orgQuota resources.OrganizationQuota) (resources.OrganizationQuota, Warnings, error) {
	var responseOrgQuota resources.OrganizationQuota

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.PostOrganizationQuotaRequest,
		RequestBody:  orgQuota,
		ResponseBody: &responseOrgQuota,
	})

	return responseOrgQuota, warnings, err
}

func (client *Client) DeleteOrganizationQuota(quotaGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := client.MakeRequest(RequestParams{
		RequestName: ccv3internal.DeleteOrganizationQuotaRequest,
		URIParams:   internal.Params{"quota_guid": quotaGUID},
	})

	return jobURL, warnings, err
}

func (client *Client) GetOrganizationQuota(quotaGUID string) (resources.OrganizationQuota, Warnings, error) {
	var responseBody resources.OrganizationQuota

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.GetOrganizationQuotaRequest,
		URIParams:    internal.Params{"quota_guid": quotaGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) GetOrganizationQuotas(query ...Query) ([]resources.OrganizationQuota, Warnings, error) {
	var organizationQuotas []resources.OrganizationQuota

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  ccv3internal.GetOrganizationQuotasRequest,
		Query:        query,
		ResponseBody: resources.OrganizationQuota{},
		AppendToList: func(item interface{}) error {
			organizationQuotas = append(organizationQuotas, item.(resources.OrganizationQuota))
			return nil
		},
	})

	return organizationQuotas, warnings, err
}

func (client *Client) UpdateOrganizationQuota(orgQuota resources.OrganizationQuota) (resources.OrganizationQuota, Warnings, error) {
	orgQuotaGUID := orgQuota.GUID
	orgQuota.GUID = ""

	var responseBody resources.OrganizationQuota

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.PatchOrganizationQuotaRequest,
		URIParams:    internal.Params{"quota_guid": orgQuotaGUID},
		RequestBody:  orgQuota,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
