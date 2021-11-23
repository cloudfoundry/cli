package configv3

type CFOnK8s struct {
	Enabled  bool   `json:"Enabled"`
	AuthInfo string `json:"AuthInfo"`
}

func (config *Config) IsCFOnK8s() bool {
	return config.ConfigFile.CFOnK8s.Enabled
}
