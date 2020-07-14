package resources

import "encoding/json"

type FeatureFlag struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func (f FeatureFlag) MarshalJSON() ([]byte, error) {
	var bodyFlag struct {
		Enabled bool `json:"enabled"`
	}

	bodyFlag.Enabled = f.Enabled

	return json.Marshal(bodyFlag)
}
