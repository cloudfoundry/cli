package config

import (
	"time"
)

type NewBrokerResponse struct {
	GUID string `json:"guid"`
}

type Plan struct {
	Name            string           `json:"name" validate:"printascii,min=5"`
	ID              string           `json:"id,omitempty"`
	Description     string           `json:"description,omitempty"`
	MaintenanceInfo *MaintenanceInfo `json:"maintenance_info,omitempty"`
	Free            bool             `json:"free"`
}

type Service struct {
	Name        string   `json:"name" validate:"printascii,min=5"`
	ID          string   `json:"id,omitempty"`
	Description string   `json:"description,omitempty"`
	Plans       []Plan   `json:"plans" validate:"min=1,dive"`
	Shareable   bool     `json:"shareable"`
	Bindable    bool     `json:"bindable"`
	Requires    []string `json:"requires,omitempty"`
}

type BrokerConfiguration struct {
	Services            []Service     `json:"services" validate:"min=1,dive"`
	Username            string        `json:"username" validate:"printascii,min=5"`
	Password            string        `json:"password" validate:"printascii,min=5"`
	CatalogResponse     int           `json:"catalog_response,omitempty" validate:"min=0,max=600"`
	ProvisionResponse   int           `json:"provision_response,omitempty" validate:"min=0,max=600"`
	UpdateResponse      int           `json:"update_response,omitempty" validate:"min=0,max=600"`
	DeprovisionResponse int           `json:"deprovision_response,omitempty" validate:"min=0,max=600"`
	BindResponse        int           `json:"bind_response,omitempty" validate:"min=0,max=600"`
	UnbindResponse      int           `json:"unbind_response,omitempty" validate:"min=0,max=600"`
	GetBindingResponse  int           `json:"get_binding_response,omitempty" validate:"min=0,max=600"`
	AsyncResponseDelay  time.Duration `json:"async_response_delay"`
}

type MaintenanceInfo struct {
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
}
