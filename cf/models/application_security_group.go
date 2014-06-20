package models

type ApplicationSecurityGroupFields struct {
	Name       string              `json:"name"`
	Guid       string              `json:"guid"`
	Rules      []map[string]string `json:"rules"`
	SpaceGuids []string            `json:"space_guids"`
}
