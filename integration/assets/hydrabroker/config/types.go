package config

type NewBrokerResponse struct {
	GUID string `json:"guid"`
}

type Plan struct {
	Name        string `json:"name" validate:"printascii,min=5"`
	GUID        string
	Description string
}

type Service struct {
	Name        string `json:"name" validate:"printascii,min=5"`
	GUID        string
	Description string
	Plans       []Plan `json:"plans" validate:"min=1,dive"`
}

type BrokerConfiguration struct {
	Services []Service `json:"services" validate:"min=1,dive"`
	Username string    `json:"username" validate:"printascii,min=5"`
	Password string    `json:"password" validate:"printascii,min=5"`
}
