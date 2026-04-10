package resources

import (
	"encoding/json"
	"time"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller"
)

type AccessRule struct {
	GUID      string     `json:"guid,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Name      string     `json:"name"`
	Selector  string     `json:"selector"`
	RouteGUID string     `json:"-"`

	// Metadata is used for custom tagging of API resources
	Metadata *Metadata `json:"metadata,omitempty"`
}

func (a AccessRule) MarshalJSON() ([]byte, error) {
	type alias struct {
		GUID      string     `json:"guid,omitempty"`
		CreatedAt *time.Time `json:"created_at,omitempty"`
		UpdatedAt *time.Time `json:"updated_at,omitempty"`
		Name      string     `json:"name"`
		Selector  string     `json:"selector"`
		Metadata  *Metadata  `json:"metadata,omitempty"`

		Relationships struct {
			Route struct {
				Data struct {
					GUID string `json:"guid"`
				} `json:"data"`
			} `json:"route"`
		} `json:"relationships"`
	}

	var aliasData alias
	aliasData.GUID = a.GUID
	aliasData.CreatedAt = a.CreatedAt
	aliasData.UpdatedAt = a.UpdatedAt
	aliasData.Name = a.Name
	aliasData.Selector = a.Selector
	aliasData.Metadata = a.Metadata
	aliasData.Relationships.Route.Data.GUID = a.RouteGUID

	return json.Marshal(aliasData)
}

func (a *AccessRule) UnmarshalJSON(data []byte) error {
	var alias struct {
		GUID      string     `json:"guid,omitempty"`
		CreatedAt *time.Time `json:"created_at,omitempty"`
		UpdatedAt *time.Time `json:"updated_at,omitempty"`
		Name      string     `json:"name"`
		Selector  string     `json:"selector"`
		Metadata  *Metadata  `json:"metadata,omitempty"`

		Relationships struct {
			Route struct {
				Data struct {
					GUID string `json:"guid,omitempty"`
				} `json:"data,omitempty"`
			} `json:"route,omitempty"`
		} `json:"relationships,omitempty"`
	}

	err := cloudcontroller.DecodeJSON(data, &alias)
	if err != nil {
		return err
	}

	a.GUID = alias.GUID
	a.CreatedAt = alias.CreatedAt
	a.UpdatedAt = alias.UpdatedAt
	a.Name = alias.Name
	a.Selector = alias.Selector
	a.RouteGUID = alias.Relationships.Route.Data.GUID
	a.Metadata = alias.Metadata

	return nil
}
