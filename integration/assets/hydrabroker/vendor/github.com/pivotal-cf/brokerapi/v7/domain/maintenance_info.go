package domain

import "reflect"

type MaintenanceInfo struct {
	Public      map[string]string `json:"public,omitempty"`
	Private     string            `json:"private,omitempty"`
	Version     string            `json:"version,omitempty"`
	Description string            `json:"description,omitempty"`
}

func (m *MaintenanceInfo) Equals(input MaintenanceInfo) bool {
	return m.Version == input.Version &&
		m.Private == input.Private &&
		reflect.DeepEqual(m.Public, input.Public)
}
