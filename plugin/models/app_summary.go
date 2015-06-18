package plugin_models

type ApplicationSummary struct {
	Name             string
	Guid             string
	State            string
	TotalInstances   int
	RunningInstances int
	Memory           int64
	DiskQuota        int64
	Routes           []RouteSummary
}
