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
	DiskQuota int64   // in bytes
	DiskUsage int64
	MemQuota  int64
	MemUsage  int64
	Host      string
	HandleId  string
}
