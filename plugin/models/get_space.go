package plugin_models

type GetSpace_Model struct {
	GetSpaces_Model
	Organization     GetSpace_Orgs
	Applications     []GetSpace_Apps
	ServiceInstances []GetSpace_ServiceInstance
	Domains          []GetSpace_Domains
	SecurityGroups   []GetSpace_SecurityGroup
	SpaceQuota       GetSpace_SpaceQuota
}

type GetSpace_Orgs struct {
	Guid string
	Name string
}

type GetSpace_Apps struct {
	Name string
	Guid string
}

type GetSpace_AppsDomainFields struct {
	Guid                   string
	Name                   string
	OwningOrganizationGuid string
	Shared                 bool
}

type GetSpace_ServiceInstance struct {
	Guid string
	Name string
}

type GetSpace_Domains struct {
	Guid                   string
	Name                   string
	OwningOrganizationGuid string
	Shared                 bool
}

type GetSpace_SecurityGroup struct {
	Name  string
	Guid  string
	Rules []map[string]interface{}
}

type GetSpace_SpaceQuota struct {
	Guid                    string
	Name                    string
	MemoryLimit             int64
	InstanceMemoryLimit     int64
	RoutesLimit             int
	ServicesLimit           int
	NonBasicServicesAllowed bool
}
