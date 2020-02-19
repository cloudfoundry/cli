package resources

type SecurityGroup struct {
	Name string `json:"name"`
	GUID string `json:"guid, omitempty"`
}
