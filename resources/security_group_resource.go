package resources

type SecurityGroup struct {
	Name  string `json:"name"`
	GUID  string `json:"guid,omitempty"`
	Rules []Rule `json:"rules"`
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
