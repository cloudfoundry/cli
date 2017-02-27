package plugin_models

import "time"

type GetAppModel struct {
	Guid                 string
	Name                 string
	BuildpackUrl         string
	Command              string
	Diego                bool
	DetectedStartCommand string
	DiskQuota            int64 // in Megabytes
	EnvironmentVars      map[string]interface{}
	InstanceCount        int
	Memory               int64 // in Megabytes
	RunningInstances     int
	HealthCheckTimeout   int
	State                string
	SpaceGuid            string
	PackageUpdatedAt     *time.Time
	PackageState         string
	StagingFailedReason  string
	AppPorts             []int
	Stack                *GetApp_Stack
	Instances            []GetApp_AppInstanceFields
	Routes               []GetApp_RouteSummary
	Services             []GetApp_ServiceSummary
}

type GetApp_AppInstanceFields struct {
	State     string
	Details   string
	Since     time.Time
	CpuUsage  float64 // percentage
	DiskQuota int64   // in bytes
	DiskUsage int64
	MemQuota  int64
	MemUsage  int64
}

type GetApp_Stack struct {
	Guid        string
	Name        string
	Description string
}

type GetApp_RouteSummary struct {
	Guid   string
	Host   string
	Domain GetApp_DomainFields
	Path   string
	Port   int
}

type GetApp_DomainFields struct {
	Guid string
	Name string
}

type GetApp_ServiceSummary struct {
	Guid string
	Name string
}
