package plugin_models

type GetOrg_Model struct {
	GUID            string
	Name            string
	QuotaDefinition QuotaFields
	Spaces          []GetOrg_Space
	Domains         []GetOrg_Domains
	SpaceQuotas     []GetOrg_SpaceQuota
}

type GetOrg_Space struct {
	GUID string
	Name string
}

type GetOrg_Domains struct {
	GUID                   string
	Name                   string
	OwningOrganizationGUID string
	Shared                 bool
}

type GetOrg_SpaceQuota struct {
	GUID                    string
	Name                    string
	MemoryLimit             int64
	InstanceMemoryLimit     int64
	RoutesLimit             int
	ServicesLimit           int
	NonBasicServicesAllowed bool
}
