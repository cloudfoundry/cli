package models

func NewQuotaFields(name string, memory uint64) (q QuotaFields) {
	q.Name = name
	q.MemoryLimit = memory
	return
}

type QuotaFields struct {
	Guid        string
	Name        string
	MemoryLimit uint64 // in Megabytes
}
