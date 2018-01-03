package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Domain represents a Cloud Controller Domain.
type Domain struct {
	GUID            string
	Name            string
	RouterGroupGUID string
	RouterGroupType constant.RouterGroupType
	Type            constant.DomainType
}

// UnmarshalJSON helps unmarshal a Cloud Controller Domain response.
func (domain *Domain) UnmarshalJSON(data []byte) error {
	var ccDomain struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name            string `json:"name"`
			RouterGroupGUID string `json:"router_group_guid"`
			RouterGroupType string `json:"router_group_type"`
		} `json:"entity"`
	}
	if err := json.Unmarshal(data, &ccDomain); err != nil {
		return err
	}

	domain.GUID = ccDomain.Metadata.GUID
	domain.Name = ccDomain.Entity.Name
	domain.RouterGroupGUID = ccDomain.Entity.RouterGroupGUID
	domain.RouterGroupType = constant.RouterGroupType(ccDomain.Entity.RouterGroupType)
	return nil
}

// GetSharedDomain returns the Shared Domain associated with the provided
// Domain GUID.
func (client *Client) GetSharedDomain(domainGUID string) (Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSharedDomainRequest,
		URIParams:   map[string]string{"shared_domain_guid": domainGUID},
	})
	if err != nil {
		return Domain{}, nil, err
	}

	var domain Domain
	response := cloudcontroller.Response{
		Result: &domain,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return Domain{}, response.Warnings, err
	}

	domain.Type = constant.SharedDomain
	return domain, response.Warnings, nil
}

// GetPrivateDomain returns the Private Domain associated with the provided
// Domain GUID.
func (client *Client) GetPrivateDomain(domainGUID string) (Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetPrivateDomainRequest,
		URIParams:   map[string]string{"private_domain_guid": domainGUID},
	})
	if err != nil {
		return Domain{}, nil, err
	}

	var domain Domain
	response := cloudcontroller.Response{
		Result: &domain,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return Domain{}, response.Warnings, err
	}

	domain.Type = constant.PrivateDomain
	return domain, response.Warnings, nil
}

// GetSharedDomains returns the global shared domains.
func (client *Client) GetSharedDomains(queries ...QQuery) ([]Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSharedDomainsRequest,
		Query:       FormatQueryParameters(queries),
	})
	if err != nil {
		return []Domain{}, nil, err
	}

	fullDomainsList := []Domain{}
	warnings, err := client.paginate(request, Domain{}, func(item interface{}) error {
		if domain, ok := item.(Domain); ok {
			domain.Type = constant.SharedDomain
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

// GetOrganizationPrivateDomains returns the private domains associated with an organization.
func (client *Client) GetOrganizationPrivateDomains(orgGUID string, queries ...QQuery) ([]Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationPrivateDomainsRequest,
		Query:       FormatQueryParameters(queries),
		URIParams:   map[string]string{"organization_guid": orgGUID},
	})
	if err != nil {
		return []Domain{}, nil, err
	}

	fullDomainsList := []Domain{}
	warnings, err := client.paginate(request, Domain{}, func(item interface{}) error {
		if domain, ok := item.(Domain); ok {
			domain.Type = constant.PrivateDomain
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
