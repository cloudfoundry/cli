package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

func (client Client) CreateSpaceQuota(spaceQuota resources.SpaceQuota) (resources.SpaceQuota, Warnings, error) {
	spaceQuotaBytes, err := json.Marshal(spaceQuota)

	if err != nil {
		return resources.SpaceQuota{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostSpaceQuotaRequest,
		Body:        bytes.NewReader(spaceQuotaBytes),
	})

	if err != nil {
		return resources.SpaceQuota{}, nil, err
	}

	var createdSpaceQuota resources.SpaceQuota
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &createdSpaceQuota,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return resources.SpaceQuota{}, response.Warnings, err
	}

	return createdSpaceQuota, response.Warnings, err
}

func (client Client) DeleteSpaceQuota(spaceQuotaGUID string) (JobURL, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteSpaceQuotaRequest,
		URIParams:   internal.Params{"quota_guid": spaceQuotaGUID},
	})
	if err != nil {
		return "", nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)
	if err != nil {
		return "", response.Warnings, err
	}

	return JobURL(response.ResourceLocationURL), response.Warnings, nil
}

func (client Client) GetSpaceQuota(spaceQuotaGUID string) (resources.SpaceQuota, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceQuotaRequest,
		URIParams:   internal.Params{"quota_guid": spaceQuotaGUID},
	})
	if err != nil {
		return resources.SpaceQuota{}, nil, err
	}
	var responseSpaceQuota resources.SpaceQuota
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseSpaceQuota,
	}
	err = client.connection.Make(request, &response)
	if err != nil {
		return resources.SpaceQuota{}, response.Warnings, err
	}

	return responseSpaceQuota, response.Warnings, nil
}

func (client *Client) GetSpaceQuotas(query ...Query) ([]resources.SpaceQuota, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceQuotasRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var spaceQuotasList []resources.SpaceQuota
	warnings, err := client.paginate(request, resources.SpaceQuota{}, func(item interface{}) error {
		if spaceQuota, ok := item.(resources.SpaceQuota); ok {
			spaceQuotasList = append(spaceQuotasList, spaceQuota)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   resources.SpaceQuota{},
				Unexpected: item,
			}
		}
		return nil
	})

	return spaceQuotasList, warnings, err
}

func (client *Client) UnsetSpaceQuota(spaceQuotaGUID, spaceGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		URIParams:   internal.Params{"quota_guid": spaceQuotaGUID, "space_guid": spaceGUID},
		RequestName: internal.DeleteSpaceQuotaFromSpaceRequest,
	})

	if err != nil {
		return nil, err
	}

	var response cloudcontroller.Response

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

func (client *Client) UpdateSpaceQuota(spaceQuota resources.SpaceQuota) (resources.SpaceQuota, Warnings, error) {
	spaceQuotaGUID := spaceQuota.GUID
	spaceQuota.GUID = ""

	quotaBytes, err := json.Marshal(spaceQuota)
	if err != nil {
		return resources.SpaceQuota{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchSpaceQuotaRequest,
		URIParams:   internal.Params{"quota_guid": spaceQuotaGUID},
		Body:        bytes.NewReader(quotaBytes),
	})
	if err != nil {
		return resources.SpaceQuota{}, nil, err
	}

	var responseSpaceQuota resources.SpaceQuota
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseSpaceQuota,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return resources.SpaceQuota{}, response.Warnings, err
	}

	return responseSpaceQuota, response.Warnings, nil
}
