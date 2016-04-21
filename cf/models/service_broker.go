package models

type ServiceBroker struct {
	GUID     string
	Name     string
	Username string
	Password string
	Url      string
	Services []ServiceOffering
}
