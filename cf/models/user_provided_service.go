package models

type UserProvidedServiceSummary struct {
	Total     int                         `json:"total_results"`
	Resources []UserProvidedServiceEntity `json:"resources"`
}

type UserProvidedService struct {
	Name           string                 `json:"name,omitempty"`
	Credentials    map[string]interface{} `json:"credentials"`
	SpaceGuid      string                 `json:"space_guid,omitempty"`
	SysLogDrainUrl string                 `json:"syslog_drain_url"`
}

type UserProvidedServiceEntity struct {
	UserProvidedService `json:"entity"`
}
