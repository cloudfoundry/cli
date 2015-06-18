package plugin_models

// represents just the attributes for an security group
type SecurityGroupFields struct {
	Name  string
	Guid  string
	Rules []map[string]interface{}
}
