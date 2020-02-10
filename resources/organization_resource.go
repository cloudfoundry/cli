package resources

import "encoding/json"

// Organization represents a Cloud Controller V3 Organization.
type Organization struct {
	// GUID is the unique organization identifier.
	GUID string `json:"guid,omitempty"`
	// Name is the name of the organization.
	Name string `json:"name"`
	// QuotaGUID is the GUID of the organization Quota applied to this Organization
	QuotaGUID string `json:"-"`

	// Metadata is used for custom tagging of API resources
	Metadata *Metadata `json:"metadata,omitempty"`
}

func (org *Organization) UnmarshalJSON(data []byte) error {
	type alias Organization
	var aliasOrg alias
	err := json.Unmarshal(data, &aliasOrg)
	if err != nil {
		return err
	}

	*org = Organization(aliasOrg)

	remainingFields := new(struct {
		Relationships struct {
			Quota struct {
				Data struct {
					GUID string
				}
			}
		}
	})

	err = json.Unmarshal(data, &remainingFields)
	if err != nil {
		return err
	}

	org.QuotaGUID = remainingFields.Relationships.Quota.Data.GUID

	return nil
}
