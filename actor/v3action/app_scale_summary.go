package v3action

// AppScaleSummary represents an application with its processes and droplet.
type AppScaleSummary struct {
	NumInstances int
	MemoryUsage  int
	DiskUsage    int
}
