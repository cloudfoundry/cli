package models

type UserProvidedService struct {
	Name           string                 `json:"name,omitempty"`
	Credentials    map[string]interface{} `json:"credentials"`
	SpaceGuid      string                 `json:"space_guid,omitempty"`
	SysLogDrainUrl string                 `json:"syslog_drain_url"`
}
