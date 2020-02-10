package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

// CheckRoute checks whether the route with the given domain GUID, hostname,
// and path exists in the foundation.
func (client Client) CheckRoute(domainGUID string, hostname string, path string) (bool, Warnings, error) {
	var query []Query

	if hostname != "" {
		query = append(query, Query{Key: HostFilter, Values: []string{hostname}})
	}

	if path != "" {
		query = append(query, Query{Key: PathFilter, Values: []string{path}})
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetDomainRouteReservationsRequest,
		URIParams:   map[string]string{"domain_guid": domainGUID},
		Query:       query,
	})
	if err != nil {
		return false, nil, err
	}

	var responseJson struct {
		MatchingRoute bool `json:"matching_route"`
	}

	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseJson,
	}
	err = client.connection.Make(request, &response)

	return responseJson.MatchingRoute, response.Warnings, err
}

func (client Client) CreateDomain(domain resources.Domain) (resources.Domain, Warnings, error) {
	bodyBytes, err := json.Marshal(domain)
	if err != nil {
		return resources.Domain{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostDomainRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return resources.Domain{}, nil, err
	}

	var ccDomain resources.Domain
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &ccDomain,
	}

	err = client.connection.Make(request, &response)

	return ccDomain, response.Warnings, err
}

func (client Client) DeleteDomain(domainGUID string) (JobURL, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		URIParams: map[string]string{
			"domain_guid": domainGUID,
		},
		RequestName: internal.DeleteDomainRequest,
	})
	if err != nil {
		return "", nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return JobURL(response.ResourceLocationURL), response.Warnings, err
}

// GetDomain returns a domain with the given GUID.
func (client *Client) GetDomain(domainGUID string) (resources.Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetDomainRequest,
		URIParams:   map[string]string{"domain_guid": domainGUID},
	})
	if err != nil {
		return resources.Domain{}, nil, err
	}

	var responseDomain resources.Domain
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseDomain,
	}
	err = client.connection.Make(request, &response)

	return responseDomain, response.Warnings, err
}

func (client Client) GetDomains(query ...Query) ([]resources.Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetDomainsRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullDomainsList []resources.Domain
	warnings, err := client.paginate(request, resources.Domain{}, func(item interface{}) error {
		if domain, ok := item.(resources.Domain); ok {
			fullDomainsList = append(fullDomainsList, domain)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   resources.Domain{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullDomainsList, warnings, err
}

func (client Client) GetOrganizationDomains(orgGUID string, query ...Query) ([]resources.Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		URIParams:   internal.Params{"organization_guid": orgGUID},
		RequestName: internal.GetOrganizationDomainsRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullDomainsList []resources.Domain
	warnings, err := client.paginate(request, resources.Domain{}, func(item interface{}) error {
		if domain, ok := item.(resources.Domain); ok {
			fullDomainsList = append(fullDomainsList, domain)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   resources.Domain{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullDomainsList, warnings, err
}

func (client Client) SharePrivateDomainToOrgs(domainGuid string, sharedOrgs resources.SharedOrgs) (Warnings, error) {
	bodyBytes, err := json.Marshal(sharedOrgs)

	if err != nil {
		return nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		URIParams:   internal.Params{"domain_guid": domainGuid},
		RequestName: internal.SharePrivateDomainRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	if err != nil {
		return nil, err
	}

	var ccSharedOrgs resources.SharedOrgs
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &ccSharedOrgs,
	}

	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

func (client Client) UnsharePrivateDomainFromOrg(domainGuid string, orgGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		URIParams:   internal.Params{"domain_guid": domainGuid, "org_guid": orgGUID},
		RequestName: internal.DeleteSharedOrgFromDomainRequest,
	})

	if err != nil {
		return nil, err
	}

	var response cloudcontroller.Response

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}
