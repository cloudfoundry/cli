package models

type ServiceBindingRequest struct {
	AppGuid             string                 `json:"app_guid"`
	ServiceInstanceGuid string                 `json:"service_instance_guid"`
	Params              map[string]interface{} `json:"parameters,omitempty"`
}

type ServiceBindingFields struct {
	Guid    string
	Url     string
	AppGuid string
}
