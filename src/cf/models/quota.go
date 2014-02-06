package models

func NewQuotaFields(name string, memory uint64) (q QuotaFields) {
	q.Name = name
	q.MemoryLimit = memory
	return
}

type QuotaFields struct {
	BasicFields
	MemoryLimit uint64 // in Megabytes
}
