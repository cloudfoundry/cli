package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type SpaceQuota struct {
	Quota
	// OrgGUID is the unique ID of the owning organization
	OrgGUID string
	// SpaceGUIDs are the list of unique ID's of the associated spaces
	SpaceGUIDs []string
}

func (sq SpaceQuota) MarshalJSON() ([]byte, error) {
	appLimits := map[string]interface{}{}
	if sq.Apps.TotalMemory != nil {
		appLimits["total_memory_in_mb"] = sq.Apps.TotalMemory
	}
	if sq.Apps.InstanceMemory != nil {
		appLimits["per_process_memory_in_mb"] = sq.Apps.InstanceMemory
	}
	if sq.Apps.TotalAppInstances != nil {
		appLimits["total_instances"] = sq.Apps.TotalAppInstances
	}

	serviceLimits := map[string]interface{}{}
	if sq.Services.PaidServicePlans != nil {
		serviceLimits["paid_services_allowed"] = sq.Services.PaidServicePlans
	}
	if sq.Services.TotalServiceInstances != nil {
		serviceLimits["total_service_instances"] = sq.Services.TotalServiceInstances
	}

	routeLimits := map[string]interface{}{}
	if sq.Routes.TotalRoutes != nil {
		routeLimits["total_routes"] = sq.Routes.TotalRoutes
	}
	if sq.Routes.TotalReservedPorts != nil {
		routeLimits["total_reserved_ports"] = sq.Routes.TotalReservedPorts
	}

	relationships := map[string]interface{}{}

	if sq.OrgGUID != "" {
		relationships["organization"] = map[string]interface{}{
			"data": map[string]interface{}{
				"guid": sq.OrgGUID,
			},
		}
	}

	if len(sq.SpaceGUIDs) > 0 {
		spaceData := make([]map[string]interface{}, len(sq.SpaceGUIDs))
		for i, spaceGUID := range sq.SpaceGUIDs {
			spaceData[i] = map[string]interface{}{
				"guid": spaceGUID,
			}
		}

		relationships["spaces"] = map[string]interface{}{
			"data": spaceData,
		}
	}

	jsonMap := map[string]interface{}{
		"name":     sq.Name,
		"apps":     appLimits,
		"services": serviceLimits,
		"routes":   routeLimits,
	}

	if len(relationships) != 0 {
		jsonMap["relationships"] = relationships
	}

	return json.Marshal(jsonMap)
}

func (sq *SpaceQuota) UnmarshalJSON(data []byte) error {
	type alias SpaceQuota
	var defaultUnmarshalledSpaceQuota alias
	err := json.Unmarshal(data, &defaultUnmarshalledSpaceQuota)
	if err != nil {
		return err
	}

	*sq = SpaceQuota(defaultUnmarshalledSpaceQuota)

	type RemainingFieldsStruct struct {
		Relationships struct {
			Organization struct {
				Data struct {
					Guid string
				}
			}
			Spaces struct {
				Data []struct {
					Guid string
				}
			}
		}
	}

	var remainingFields RemainingFieldsStruct
	err = json.Unmarshal(data, &remainingFields)
	if err != nil {
		return err
	}

	sq.OrgGUID = remainingFields.Relationships.Organization.Data.Guid

	for _, spaceData := range remainingFields.Relationships.Spaces.Data {
		sq.SpaceGUIDs = append(sq.SpaceGUIDs, spaceData.Guid)
	}

	return nil
}

func (client Client) ApplySpaceQuota(quotaGUID string, spaceGUID string) (RelationshipList, Warnings, error) {
	relationshipList := RelationshipList{GUIDs: []string{spaceGUID}}
	relationshipListBytes, err := json.Marshal(relationshipList)
	if err != nil {
		return RelationshipList{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostSpaceQuotaRelationshipsRequest,
		URIParams:   internal.Params{"quota_guid": quotaGUID},
		Body:        bytes.NewReader(relationshipListBytes),
	})
	if err != nil {
		return RelationshipList{}, nil, err
	}

	var appliedRelationshipList RelationshipList
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &appliedRelationshipList,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return RelationshipList{}, response.Warnings, err
	}

	return appliedRelationshipList, response.Warnings, nil
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

	var createdSpaceQuota SpaceQuota
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &createdSpaceQuota,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return SpaceQuota{}, response.Warnings, err
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

func (client Client) GetSpaceQuota(spaceQuotaGUID string) (SpaceQuota, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceQuotaRequest,
		URIParams:   internal.Params{"quota_guid": spaceQuotaGUID},
	})
	if err != nil {
		return SpaceQuota{}, nil, err
	}
	var responseSpaceQuota SpaceQuota
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseSpaceQuota,
	}
	err = client.connection.Make(request, &response)
	if err != nil {
		return SpaceQuota{}, response.Warnings, err
	}

	return responseSpaceQuota, response.Warnings, nil
}

func (client *Client) GetSpaceQuotas(query ...Query) ([]SpaceQuota, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceQuotasRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var spaceQuotasList []SpaceQuota
	warnings, err := client.paginate(request, SpaceQuota{}, func(item interface{}) error {
		if spaceQuota, ok := item.(SpaceQuota); ok {
			spaceQuotasList = append(spaceQuotasList, spaceQuota)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   SpaceQuota{},
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

func (client *Client) UpdateSpaceQuota(spaceQuota SpaceQuota) (SpaceQuota, Warnings, error) {
	spaceQuotaGUID := spaceQuota.GUID
	spaceQuota.GUID = ""

	quotaBytes, err := json.Marshal(spaceQuota)
	if err != nil {
		return SpaceQuota{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchSpaceQuotaRequest,
		URIParams:   internal.Params{"quota_guid": spaceQuotaGUID},
		Body:        bytes.NewReader(quotaBytes),
	})
	if err != nil {
		return SpaceQuota{}, nil, err
	}

	var responseSpaceQuota SpaceQuota
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseSpaceQuota,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return SpaceQuota{}, response.Warnings, err
	}

	return responseSpaceQuota, response.Warnings, nil
}
