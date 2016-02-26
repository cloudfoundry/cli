package plugin_models

type GetAppsModel struct {
	Name             string
	Guid             string
	State            string
	TotalInstances   int
	RunningInstances int
	Memory           int64
	DiskQuota        int64
	Routes           []GetAppsRouteSummary
	AppPorts         []int
}

type GetAppsRouteSummary struct {
	Guid   string
	Host   string
	Domain GetAppsDomainFields
}

type GetAppsDomainFields struct {
	Guid                   string
	Name                   string
	OwningOrganizationGuid string
	Shared                 bool
}
