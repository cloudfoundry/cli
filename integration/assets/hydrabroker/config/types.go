package config

type NewBrokerResponse struct {
	GUID string `json:"guid"`
}

type Plan struct {
	Name        string `json:"name" validate:"printascii,min=5"`
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
}

type Service struct {
	Name        string `json:"name" validate:"printascii,min=5"`
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
	Plans       []Plan `json:"plans" validate:"min=1,dive"`
}

type BrokerConfiguration struct {
	Services            []Service `json:"services" validate:"min=1,dive"`
	Username            string    `json:"username" validate:"printascii,min=5"`
	Password            string    `json:"password" validate:"printascii,min=5"`
	CatalogResponse     int       `json:"catalog_response,omitempty" validate:"min=0,max=600"`
	ProvisionResponse   int       `json:"provision_response,omitempty" validate:"min=0,max=600"`
	DeprovisionResponse int       `json:"deprovision_response,omitempty" validate:"min=0,max=600"`
}
