package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type SpaceQuota struct {
	// GUID is the unique ID of the space quota.
	GUID string
	// Name is the name of the space quota
	Name string
	// Apps contain the various limits that are associated with applications
	Apps AppLimit
	// Services contain the various limits that are associated with services
	Services ServiceLimit
	// Routes contain the various limits that are associated with routes
	Routes RouteLimit
	// OrgGUID is the unique ID of the owning organization
	OrgGUID string
	// SpaceGUIDs are the list of unique ID's of the associated spaces
	SpaceGUIDs []string
}

func (sq SpaceQuota) MarshalJSON() ([]byte, error) {
	appsStruct := map[string]interface{}{
		"total_memory_in_mb":       sq.Apps.TotalMemory,
		"per_process_memory_in_mb": sq.Apps.InstanceMemory,
		"total_instances":          sq.Apps.TotalAppInstances,
	}

	servicesSruct := map[string]interface{}{
		"paid_services_allowed":   sq.Services.PaidServicePlans,
		"total_service_instances": sq.Services.TotalServiceInstances,
	}
	routesStruct := map[string]interface{}{
		"total_routes":         sq.Routes.TotalRoutes,
		"total_reserved_ports": sq.Routes.TotalReservedPorts,
	}

	relationshipsStruct := map[string]interface{}{
		"organization": map[string]interface{}{
			"data": map[string]interface{}{
				"guid": sq.OrgGUID,
			},
		},
	}

	if len(sq.SpaceGUIDs) > 0 {
		spaceData := make([]map[string]interface{}, len(sq.SpaceGUIDs))
		for i, spaceGUID := range sq.SpaceGUIDs {
			spaceData[i] = map[string]interface{}{
				"guid": spaceGUID,
			}
		}

		relationshipsStruct["spaces"] = map[string]interface{}{
			"data": spaceData,
		}
	}

	jsonMap := map[string]interface{}{
		"name":          sq.Name,
		"apps":          appsStruct,
		"services":      servicesSruct,
		"routes":        routesStruct,
		"relationships": relationshipsStruct,
	}

	return json.Marshal(jsonMap)
}

func (sq *SpaceQuota) UnmarshalJSON(data []byte) error {
	type spaceQuotaWithoutCustomUnmarshal SpaceQuota
	var defaultUnmarshalledSpaceQuota spaceQuotaWithoutCustomUnmarshal
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
