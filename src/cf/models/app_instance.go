package models

import "time"

type InstanceState string

const (
	InstanceStarting InstanceState = "starting"
	InstanceRunning  InstanceState = "running"
	InstanceFlapping InstanceState = "flapping"
	InstanceDown     InstanceState = "down"
)

type AppInstanceFields struct {
	State     InstanceState
	Since     time.Time
	CpuUsage  float64 // percentage
	DiskQuota uint64  // in bytes
	DiskUsage uint64
	MemQuota  uint64
	MemUsage  uint64
}
