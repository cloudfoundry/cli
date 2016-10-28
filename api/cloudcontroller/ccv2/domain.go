package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Domain represents a Cloud Controller Domain.
type Domain struct {
	GUID string
	Name string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Domain response.
func (domain *Domain) UnmarshalJSON(data []byte) error {
	var ccDomain struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name string `json:"name"`
		} `json:"entity"`
	}
	if err := json.Unmarshal(data, &ccDomain); err != nil {
		return err
	}

	domain.GUID = ccDomain.Metadata.GUID
	domain.Name = ccDomain.Entity.Name
	return nil
}

// GetSharedDomain returns the Shared Domain associated with the provided
// Domain GUID.
func (client *Client) GetSharedDomain(domainGUID string) (Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.SharedDomainRequest,
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

	return domain, response.Warnings, nil
}

// GetPrivateDomain returns the Private Domain associated with the provided
// Domain GUID.
func (client *Client) GetPrivateDomain(domainGUID string) (Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PrivateDomainRequest,
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

	return domain, response.Warnings, nil
}
