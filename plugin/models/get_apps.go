package plugin_models

type GetAppsModel struct {
	Name             string
	GUID             string
	State            string
	TotalInstances   int
	RunningInstances int
	Memory           int64
	DiskQuota        int64
	Routes           []GetAppsRouteSummary
	AppPorts         []int
}

type GetAppsRouteSummary struct {
	GUID   string
	Host   string
	Domain GetAppsDomainFields
}

type GetAppsDomainFields struct {
	GUID                   string
	Name                   string
	OwningOrganizationGUID string
	Shared                 bool
}
