package plugin_models

type GetOrgUsers_Model struct {
	GUID     string
	Username string
	IsAdmin  bool
	Roles    []string
}
