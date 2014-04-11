package models

func NewQuotaFields(name string, memory uint64) (q QuotaFields) {
	q.Name = name
	q.MemoryLimit = memory
	return
}

type QuotaFields struct {
	Guid                    string `json:"guid,omitempty"`
	NonBasicServicesAllowed bool   `json:"non_basic_services_allowed"`
	Name                    string `json:"name"`
	MemoryLimit             uint64 `json:"memory_limit"` // in Megabytes
	RoutesLimit             uint   `json:"total_routes"`
	ServicesLimit           uint   `json:"total_services"`
}
