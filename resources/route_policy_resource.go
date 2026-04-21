package resources

import (
	"encoding/json"
	"time"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller"
)

type RoutePolicy struct {
	GUID      string     `json:"guid,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Source    string     `json:"source"`
	RouteGUID string     `json:"-"`

	// Metadata is used for custom tagging of API resources
	Metadata *Metadata `json:"metadata,omitempty"`
}

func (r RoutePolicy) MarshalJSON() ([]byte, error) {
	type alias struct {
		GUID      string     `json:"guid,omitempty"`
		CreatedAt *time.Time `json:"created_at,omitempty"`
		UpdatedAt *time.Time `json:"updated_at,omitempty"`
		Source    string     `json:"source"`
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
	aliasData.GUID = r.GUID
	aliasData.CreatedAt = r.CreatedAt
	aliasData.UpdatedAt = r.UpdatedAt
	aliasData.Source = r.Source
	aliasData.Metadata = r.Metadata
	aliasData.Relationships.Route.Data.GUID = r.RouteGUID

	return json.Marshal(aliasData)
}

func (r *RoutePolicy) UnmarshalJSON(data []byte) error {
	var alias struct {
		GUID      string     `json:"guid,omitempty"`
		CreatedAt *time.Time `json:"created_at,omitempty"`
		UpdatedAt *time.Time `json:"updated_at,omitempty"`
		Source    string     `json:"source"`
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

	r.GUID = alias.GUID
	r.CreatedAt = alias.CreatedAt
	r.UpdatedAt = alias.UpdatedAt
	r.Source = alias.Source
	r.RouteGUID = alias.Relationships.Route.Data.GUID
	r.Metadata = alias.Metadata

	return nil
}
