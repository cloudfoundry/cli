package models

func NewEnvironment() *Environment {
	return &Environment{
		System:      make(map[string]interface{}),
		Environment: make(map[string]string),
		Running:     make(map[string]string),
		Staging:     make(map[string]string),
	}
}

type Environment struct {
	System      map[string]interface{} `json:"system_env_json,omitempty"`
	Environment map[string]string      `json:"environment_json,omitempty"`
	Running     map[string]string      `json:"running_env_json,omitempty"`
	Staging     map[string]string      `json:"staging_env_json,omitempty"`
}
