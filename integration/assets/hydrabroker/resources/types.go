package resources

import (
	"encoding/json"
)

// JSONObject represents a JSON object. When not initialized it serializes to `{}`,
// whereas a `map[string]interface{}` serializes to `null`.
type JSONObject map[string]interface{}

func (j JSONObject) MarshalJSON() ([]byte, error) {
	switch len(j) {
	case 0:
		return []byte("{}"), nil
	default:
		return json.Marshal(map[string]interface{}(j))
	}
}

type ServiceInstanceDetails struct {
	ServiceID  string     `json:"service_id"`
	PlanID     string     `json:"plan_id"`
	Parameters JSONObject `json:"parameters"`
}

type BindingDetails struct {
	Parameters  JSONObject `json:"parameters"`
	Credentials JSONObject `json:"credentials"`
}
