package models

func NewFeatureFlag(name string, enabled bool, errorMessage string) (f FeatureFlag) {
	f.Name = name
	f.Enabled = enabled
	f.ErrorMessage = errorMessage
	return
}

type FeatureFlag struct {
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	ErrorMessage string `json:"error_message"`
}
