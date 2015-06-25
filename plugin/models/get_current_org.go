package plugin_models

type Organization struct {
	OrganizationFields
}

type OrganizationFields struct {
	Guid            string
	Name            string
	QuotaDefinition QuotaFields
}

type QuotaFields struct {
	Guid                    string
	Name                    string
	MemoryLimit             int64
	InstanceMemoryLimit     int64
	RoutesLimit             int
	ServicesLimit           int
	NonBasicServicesAllowed bool
}
