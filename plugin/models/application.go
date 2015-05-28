package plugin_models

import "time"

type Application struct {
	ApplicationFields
	Stack    *Stack
	Routes   []RouteSummary
	Services []ServicePlanSummary
}

type ApplicationFields struct {
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
	Instances            []AppInstanceFields
}

type AppInstanceFields struct {
	State     string
	Details   string
	Since     time.Time
	CpuUsage  float64 // percentage
	DiskQuota int64   // in bytes
	DiskUsage int64
	MemQuota  int64
	MemUsage  int64
}

type Stack struct {
	Guid        string
	Name        string
	Description string
}
