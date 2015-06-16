package plugin_models

type User struct {
	Guid     string
	Username string
	IsAdmin  bool
	Roles    []string
}
