package models

type AppStatsUsageFields struct {
	Cpu  float64
	Disk uint64
	Mem  uint64
	Time string
}

type AppStatsStatsFields struct {
	DiskQuota uint64 `json:"disk_quota"`
	MemQuota  uint64 `json:"mem_quota"`
	Uptime    uint64
	Usage     AppStatsUsageFields
}

type AppStatsFields struct {
	State InstanceState
	Stats AppStatsStatsFields
}
