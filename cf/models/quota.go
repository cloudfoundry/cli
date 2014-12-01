package models

func NewQuotaFields(name string, memory int64, InstanceMemoryLimit int64, routes int, services int, nonbasicservices bool) (q QuotaFields) {
	q.Name = name
	q.MemoryLimit = memory
	q.InstanceMemoryLimit = InstanceMemoryLimit
	q.RoutesLimit = routes
	q.ServicesLimit = services
	q.NonBasicServicesAllowed = nonbasicservices
	return
}

type QuotaFields struct {
	Guid                    string `json:"guid,omitempty"`
	Name                    string `json:"name"`
	MemoryLimit             int64  `json:"memory_limit"`          // in Megabytes
	InstanceMemoryLimit     int64  `json:"instance_memory_limit"` // in Megabytes
	RoutesLimit             int    `json:"total_routes"`
	ServicesLimit           int    `json:"total_services"`
	NonBasicServicesAllowed bool   `json:"non_basic_services_allowed"`
}
