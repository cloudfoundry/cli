package models

type ServiceKeyFields struct {
	Name                string
	Guid                string
	Url                 string
	ServiceInstanceGuid string
	ServiceInstanceUrl  string
}

type ServiceKey struct {
	Fields      ServiceKeyFields
	Credentials map[string]interface{}
}
