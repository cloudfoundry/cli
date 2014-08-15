package models

func NewSpaceQuota(name string, memory uint64, routes int, services int, nonbasicservices bool, orgGuid string) (q SpaceQuota) {
	q.Name = name
	q.MemoryLimit = memory
	q.RoutesLimit = routes
	q.ServicesLimit = services
	q.NonBasicServicesAllowed = nonbasicservices
	q.OrgGuid = orgGuid
	return
}

type SpaceQuota struct {
	Guid                    string `json:"guid,omitempty"`
	Name                    string `json:"name"`
	MemoryLimit             uint64 `json:"memory_limit"`                    // in Megabytes
	InstanceMemoryLimit     uint64 `json:"instance_memory_limit,omitempty"` // in Megabytes
	RoutesLimit             int    `json:"total_routes"`
	ServicesLimit           int    `json:"total_services"`
	NonBasicServicesAllowed bool   `json:"non_basic_services_allowed"`
	OrgGuid                 string `json:"organization_guid"`
}
