package models

// represents just the attributes for an application security group
type SecurityGroupFields struct {
	Name  string
	Guid  string
	Rules []map[string]string
}

// represents the JSON that we send up to CC when the user creates / updates a record
type SecurityGroupParams struct {
	Name       string              `json:"name"`
	Guid       string              `json:"guid,omitempty"`
	Rules      []map[string]string `json:"rules"`
	SpaceGuids []string            `json:"space_guids"`
}

// represents a fully instantiated model returned by the CC (e.g.: with its attributes and the fields for its child objects)
type SecurityGroup struct {
	SecurityGroupFields
	Spaces []Space
}
