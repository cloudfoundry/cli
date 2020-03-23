package resources

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller/jsonry"
)

type SecurityGroup struct {
	Name                   string   `jsonry:"name,omitempty"`
	GUID                   string   `jsonry:"guid,omitempty"`
	Rules                  []Rule   `jsonry:"rules,omitempty"`
	StagingGloballyEnabled *bool    `jsonry:"globally_enabled.staging,omitempty"`
	RunningGloballyEnabled *bool    `jsonry:"globally_enabled.running,omitempty"`
	StagingSpaceGUIDs      []string `jsonry:"relationships.staging_spaces.data[].guid,omitempty"`
	RunningSpaceGUIDs      []string `jsonry:"relationships.running_spaces.data[].guid,omitempty"`
}

func (sg SecurityGroup) MarshalJSON() ([]byte, error) {
	return jsonry.Marshal(sg)
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
			Staging *bool
			Running *bool
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
