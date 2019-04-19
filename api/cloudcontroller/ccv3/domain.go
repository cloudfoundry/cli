package ccv3

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
	"encoding/json"
)

type Domain struct {
	GUID             string         `json:"guid,omitempty"`
	Name             string         `json:"name"`
	Internal         types.NullBool `json:"internal,omitempty"`
	OrganizationGuid string         `json:"orgguid,omitempty"`
}

func (d Domain) MarshalJSON() ([]byte, error) {
	ccDom := ccDomain{
		Name: d.Name,
	}

	if d.Internal.IsSet {
		ccDom.Internal = &d.Internal.Value
	}

	if d.GUID != "" {
		ccDom.GUID = d.GUID
	}

	if d.OrganizationGuid != "" {
		ccDom.Relationships = &OrgRelationship{OrgData{Data{GUID: d.OrganizationGuid}}}
	}
	return json.Marshal(ccDom)
}

func (d *Domain) UnmarshalJSON(data []byte) error {
	var alias struct {
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
	}

	err := cloudcontroller.DecodeJSON(data, &alias)
	if err != nil {
		return err
	}
	d.GUID = alias.GUID
	d.Name = alias.Name
	d.Internal = alias.Internal
	d.OrganizationGuid = alias.Relationships.Organization.Data.GUID
	return nil
}

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

func (client Client) CreateDomain(domain Domain) (Domain, Warnings, error) {
	bodyBytes, err := json.Marshal(domain)
	if err != nil {
		return Domain{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostDomainRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return Domain{}, nil, err
	}

	var ccDomain Domain
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &ccDomain,
	}

	err = client.connection.Make(request, &response)

	return ccDomain, response.Warnings, err
}

func (client Client) GetDomains(query ...Query) ([]Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetDomainsRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullDomainsList []Domain
	warnings, err := client.paginate(request, Domain{}, func(item interface{}) error {
		if domain, ok := item.(Domain); ok {
			fullDomainsList = append(fullDomainsList, domain)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Domain{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullDomainsList, warnings, err
}
