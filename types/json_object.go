package types

import "encoding/json"

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
