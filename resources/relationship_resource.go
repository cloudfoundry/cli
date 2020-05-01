package resources

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

// Relationships represent associations between resources. Relationships is a
// map of RelationshipTypes to Relationship.
type Relationships map[constant.RelationshipType]Relationship

// Relationship represents a one to one relationship.
// An empty GUID will be marshaled as `null`.
type Relationship struct {
	GUID string
}

func (r Relationship) MarshalJSON() ([]byte, error) {
	if r.GUID == "" {
		var emptyCCRelationship struct {
			Data interface{} `json:"data"`
		}
		return json.Marshal(emptyCCRelationship)
	}

	var ccRelationship struct {
		Data struct {
			GUID string `json:"guid"`
		} `json:"data"`
	}

	ccRelationship.Data.GUID = r.GUID
	return json.Marshal(ccRelationship)
}

func (r *Relationship) UnmarshalJSON(data []byte) error {
	var ccRelationship struct {
		Data struct {
			GUID string `json:"guid"`
		} `json:"data"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccRelationship)
	if err != nil {
		return err
	}

	r.GUID = ccRelationship.Data.GUID
	return nil
}

// RelationshipList represents a one to many relationship.
type RelationshipList struct {
	GUIDs []string
}

func (r RelationshipList) MarshalJSON() ([]byte, error) {
	var ccRelationship struct {
		Data []map[string]string `json:"data"`
	}

	for _, guid := range r.GUIDs {
		ccRelationship.Data = append(
			ccRelationship.Data,
			map[string]string{
				"guid": guid,
			})
	}

	return json.Marshal(ccRelationship)
}

func (r *RelationshipList) UnmarshalJSON(data []byte) error {
	var ccRelationships struct {
		Data []map[string]string `json:"data"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccRelationships)
	if err != nil {
		return err
	}

	for _, partner := range ccRelationships.Data {
		r.GUIDs = append(r.GUIDs, partner["guid"])
	}
	return nil
}
