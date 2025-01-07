package resources

import (
	"code.cloudfoundry.org/cli/v8/types"
	"code.cloudfoundry.org/jsonry"
)

type Domain struct {
	GUID             string         `json:"guid,omitempty"`
	Name             string         `json:"name"`
	Internal         types.NullBool `json:"internal,omitempty"`
	OrganizationGUID string         `jsonry:"relationships.organization.data.guid,omitempty"`
	RouterGroup      string         `jsonry:"router_group.guid,omitempty"`
	Protocols        []string       `jsonry:"supported_protocols,omitempty"`

	// Metadata is used for custom tagging of API resources
	Metadata *Metadata `json:"metadata,omitempty"`
}

func (d Domain) MarshalJSON() ([]byte, error) {
	type domainWithBoolPointer struct {
		GUID             string   `jsonry:"guid,omitempty"`
		Name             string   `jsonry:"name"`
		Internal         *bool    `jsonry:"internal,omitempty"`
		OrganizationGUID string   `jsonry:"relationships.organization.data.guid,omitempty"`
		RouterGroup      string   `jsonry:"router_group.guid,omitempty"`
		Protocols        []string `jsonry:"supported_protocols,omitempty"`
	}

	clone := domainWithBoolPointer{
		GUID:             d.GUID,
		Name:             d.Name,
		OrganizationGUID: d.OrganizationGUID,
		RouterGroup:      d.RouterGroup,
		Protocols:        d.Protocols,
	}

	if d.Internal.IsSet {
		clone.Internal = &d.Internal.Value
	}
	return jsonry.Marshal(clone)
}

func (d *Domain) UnmarshalJSON(data []byte) error {
	type alias Domain
	var defaultUnmarshalledDomain alias
	err := jsonry.Unmarshal(data, &defaultUnmarshalledDomain)
	if err != nil {
		return err
	}

	*d = Domain(defaultUnmarshalledDomain)

	return nil
}

func (d Domain) Shared() bool {
	return d.OrganizationGUID == ""
}

func (d *Domain) IsTCP() bool {
	for _, p := range d.Protocols {
		if p == "tcp" {
			return true
		}
	}

	return false
}
