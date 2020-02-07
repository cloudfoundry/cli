package resources

import "encoding/json"

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
