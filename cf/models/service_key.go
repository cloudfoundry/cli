package models

type ServiceKeyFields struct {
	Name                string
	Guid                string
	Url                 string
	ServiceInstanceGuid string
	ServiceInstanceUrl  string
}

type ServiceKeyRequest struct {
	Name                string                 `json:"name"`
	ServiceInstanceGuid string                 `json:"service_instance_guid"`
	Params              map[string]interface{} `json:"parameters,omitempty"`
}

type ServiceKey struct {
	Fields      ServiceKeyFields
	Credentials map[string]interface{}
}
