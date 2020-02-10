package resources

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/types"
)

type Domain struct {
	GUID             string         `json:"guid,omitempty"`
	Name             string         `json:"name"`
	Internal         types.NullBool `json:"internal,omitempty"`
	OrganizationGUID string         `json:"orgguid,omitempty"`

	// Metadata is used for custom tagging of API resources
	Metadata *Metadata `json:"metadata,omitempty"`
}

func (d Domain) MarshalJSON() ([]byte, error) {
	type Data struct {
		GUID string `json:"guid,omitempty"`
	}

	type OrgData struct {
		Data Data `json:"data,omitempty"`
	}

	type OrgRelationship struct {
		Org OrgData `json:"organization,omitempty"`
	}

	type ccDomain struct {
		GUID          string           `json:"guid,omitempty"`
		Name          string           `json:"name"`
		Internal      *bool            `json:"internal,omitempty"`
		Relationships *OrgRelationship `json:"relationships,omitempty"`
	}

	ccDom := ccDomain{
		Name: d.Name,
	}

	if d.Internal.IsSet {
		ccDom.Internal = &d.Internal.Value
	}

	if d.GUID != "" {
		ccDom.GUID = d.GUID
	}

	if d.OrganizationGUID != "" {
		ccDom.Relationships = &OrgRelationship{OrgData{Data{GUID: d.OrganizationGUID}}}
	}
	return json.Marshal(ccDom)
}

func (d *Domain) UnmarshalJSON(data []byte) error {
	var ccRouteStruct struct {
		GUID          string         `json:"guid,omitempty"`
		Name          string         `json:"name"`
		Internal      types.NullBool `json:"internal,omitempty"`
		Relationships struct {
			Organization struct {
				Data struct {
					GUID string `json:"guid,omitempty"`
				} `json:"data,omitempty"`
			} `json:"organization,omitempty"`
		} `json:"relationships,omitempty"`
		Metadata *Metadata
	}

	err := cloudcontroller.DecodeJSON(data, &ccRouteStruct)
	if err != nil {
		return err
	}

	d.GUID = ccRouteStruct.GUID
	d.Name = ccRouteStruct.Name
	d.Internal = ccRouteStruct.Internal
	d.OrganizationGUID = ccRouteStruct.Relationships.Organization.Data.GUID
	d.Metadata = ccRouteStruct.Metadata
	return nil
}

func (d Domain) Shared() bool {
	return d.OrganizationGUID == ""
}

type SharedOrgs struct {
	GUIDs []string
}

func (sharedOrgs SharedOrgs) MarshalJSON() ([]byte, error) {
	type Org struct {
		GUID string `json:"guid,omitempty"`
	}

	type Data = []Org

	type sharedOrgsRelationship struct {
		Data Data `json:"data"`
	}

	var orgs []Org
	for _, sharedOrgGUID := range sharedOrgs.GUIDs {
		orgs = append(orgs, Org{GUID: sharedOrgGUID})
	}

	relationship := sharedOrgsRelationship{
		Data: orgs,
	}

	return json.Marshal(relationship)
}

func (sharedOrgs *SharedOrgs) UnmarshalJSON(data []byte) error {
	var alias struct {
		Data []struct {
			GUID string `json:"guid,omitempty"`
		} `json:"data,omitempty"`
	}

	err := cloudcontroller.DecodeJSON(data, &alias)
	if err != nil {
		return err
	}

	var guids []string
	for _, org := range alias.Data {
		guids = append(guids, org.GUID)
	}

	sharedOrgs.GUIDs = guids
	return nil
}
