package plugin_models

type GetOrg_Model struct {
	Guid            string
	Name            string
	QuotaDefinition QuotaFields
	Spaces          []GetOrg_Space
	Domains         []GetOrg_Domains
	SpaceQuotas     []GetOrg_SpaceQuota
}

type GetOrg_Space struct {
	Guid string
	Name string
}

type GetOrg_Domains struct {
	Guid                   string
	Name                   string
	OwningOrganizationGuid string
	Shared                 bool
}

type GetOrg_SpaceQuota struct {
	Guid                    string
	Name                    string
	MemoryLimit             int64
	InstanceMemoryLimit     int64
	RoutesLimit             int
	ServicesLimit           int
	NonBasicServicesAllowed bool
}
