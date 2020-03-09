package resources

import (
	"encoding/json"
)

type SecurityGroup struct {
	Name                   string `json:"name"`
	GUID                   string `json:"guid,omitempty"`
	Rules                  []Rule `json:"rules"`
	StagingGloballyEnabled bool
	RunningGloballyEnabled bool
	StagingSpaceGUIDs      []string
	RunningSpaceGUIDs      []string
}

func (sg SecurityGroup) MarshalJSON() ([]byte, error) {
	relationships := map[string]interface{}{}
	if len(sg.StagingSpaceGUIDs) > 0 {
		stagingSpaceData := make([]map[string]interface{}, len(sg.StagingSpaceGUIDs))
		for i, spaceGUID := range sg.StagingSpaceGUIDs {
			stagingSpaceData[i] = map[string]interface{}{
				"guid": spaceGUID,
			}
		}

		relationships["staging_spaces"] = map[string]interface{}{
			"data": stagingSpaceData,
		}
	}

	if len(sg.RunningSpaceGUIDs) > 0 {
		runningSpaceData := make([]map[string]interface{}, len(sg.RunningSpaceGUIDs))
		for i, spaceGUID := range sg.RunningSpaceGUIDs {
			runningSpaceData[i] = map[string]interface{}{
				"guid": spaceGUID,
			}
		}

		relationships["running_spaces"] = map[string]interface{}{
			"data": runningSpaceData,
		}
	}

	var rules = make([]Rule, 0)
	if sg.Rules != nil {
		rules = sg.Rules
	}

	globallyEnabled := map[string]bool{
		"running": sg.RunningGloballyEnabled,
		"staging": sg.StagingGloballyEnabled,
	}

	jsonMap := map[string]interface{}{
		"name":             sg.Name,
		"rules":            rules,
		"globally_enabled": globallyEnabled,
	}

	if sg.GUID != "" {
		jsonMap["guid"] = sg.GUID
	}

	if len(relationships) != 0 {
		jsonMap["relationships"] = relationships
	}

	return json.Marshal(jsonMap)
}

func (sg *SecurityGroup) UnmarshalJSON(data []byte) error {
	type alias SecurityGroup
	var defaultUnmarshalledSecurityGroup alias
	err := json.Unmarshal(data, &defaultUnmarshalledSecurityGroup)
	if err != nil {
		return err
	}

	*sg = SecurityGroup(defaultUnmarshalledSecurityGroup)

	type RemainingFieldsStruct struct {
		GloballyEnabled struct {
			Staging bool
			Running bool
		} `json:"globally_enabled"`
		Relationships struct {
			StagingSpaces struct {
				Data []struct {
					Guid string
				}
			} `json:"staging_spaces"`
			RunningSpaces struct {
				Data []struct {
					Guid string
				}
			} `json:"running_spaces"`
		}
	}

	var remainingFields RemainingFieldsStruct
	err = json.Unmarshal(data, &remainingFields)
	if err != nil {
		return err
	}

	for _, stagingSpace := range remainingFields.Relationships.StagingSpaces.Data {
		sg.StagingSpaceGUIDs = append(sg.StagingSpaceGUIDs, stagingSpace.Guid)
	}

	for _, runningSpace := range remainingFields.Relationships.RunningSpaces.Data {
		sg.RunningSpaceGUIDs = append(sg.RunningSpaceGUIDs, runningSpace.Guid)
	}

	sg.StagingGloballyEnabled = remainingFields.GloballyEnabled.Staging
	sg.RunningGloballyEnabled = remainingFields.GloballyEnabled.Running

	return nil
}

type Rule struct {
	Protocol    string  `json:"protocol"`
	Destination string  `json:"destination"`
	Ports       *string `json:"ports,omitempty"`
	Type        *int    `json:"type,omitempty"`
	Code        *int    `json:"code,omitempty"`
	Description *string `json:"description,omitempty"`
	Log         *bool   `json:"log,omitempty"`
}
