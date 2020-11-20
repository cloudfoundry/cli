package resources

import "encoding/json"

// FeatureFlag represents a Cloud Controller V3 Feature Flag.
type FeatureFlag struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func (f FeatureFlag) MarshalJSON() ([]byte, error) {
	var ccBodyFlag struct {
		Enabled bool `json:"enabled"`
	}

	ccBodyFlag.Enabled = f.Enabled

	return json.Marshal(ccBodyFlag)
}
