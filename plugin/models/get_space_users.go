package plugin_models

type GetSpaceUsers_Model struct {
	GUID     string
	Username string
	IsAdmin  bool
	Roles    []string
}
