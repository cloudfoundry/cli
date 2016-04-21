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
	GUID string
	Name string
}

type GetSpace_Apps struct {
	Name string
	GUID string
}

type GetSpace_AppsDomainFields struct {
	GUID                   string
	Name                   string
	OwningOrganizationGUID string
	Shared                 bool
}

type GetSpace_ServiceInstance struct {
	GUID string
	Name string
}

type GetSpace_Domains struct {
	GUID                   string
	Name                   string
	OwningOrganizationGUID string
	Shared                 bool
}

type GetSpace_SecurityGroup struct {
	Name  string
	GUID  string
	Rules []map[string]interface{}
}

type GetSpace_SpaceQuota struct {
	GUID                    string
	Name                    string
	MemoryLimit             int64
	InstanceMemoryLimit     int64
	RoutesLimit             int
	ServicesLimit           int
	NonBasicServicesAllowed bool
}
