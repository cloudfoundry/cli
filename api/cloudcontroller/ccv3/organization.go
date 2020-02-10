package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

// CreateOrganization creates an organization with the given name.
func (client *Client) CreateOrganization(orgName string) (resources.Organization, Warnings, error) {
	org := resources.Organization{Name: orgName}
	orgBytes, err := json.Marshal(org)
	if err != nil {
		return resources.Organization{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostOrganizationRequest,
		Body:        bytes.NewReader(orgBytes),
	})

	if err != nil {
		return resources.Organization{}, nil, err
	}

	var responseOrg resources.Organization
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseOrg,
	}
	err = client.connection.Make(request, &response)

	if err != nil {
		return resources.Organization{}, response.Warnings, err
	}

	return responseOrg, response.Warnings, err
}

// DeleteOrganization deletes the organization with the given GUID.
func (client *Client) DeleteOrganization(orgGUID string) (JobURL, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteOrganizationRequest,
		URIParams:   map[string]string{"organization_guid": orgGUID},
	})

	if err != nil {
		return "", nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)
	return JobURL(response.ResourceLocationURL), response.Warnings, err
}

// GetDefaultDomain gets the default domain for the organization with the given GUID.
func (client *Client) GetDefaultDomain(orgGUID string) (resources.Domain, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetDefaultDomainRequest,
		URIParams:   map[string]string{"organization_guid": orgGUID},
	})
	if err != nil {
		return resources.Domain{}, nil, err
	}

	var defaultDomain resources.Domain

	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &defaultDomain,
	}

	err = client.connection.Make(request, &response)

	return defaultDomain, response.Warnings, err
}

// GetIsolationSegmentOrganizations lists organizations
// entitled to an isolation segment.
func (client *Client) GetIsolationSegmentOrganizations(isolationSegmentGUID string) ([]resources.Organization, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetIsolationSegmentOrganizationsRequest,
		URIParams:   map[string]string{"isolation_segment_guid": isolationSegmentGUID},
	})
	if err != nil {
		return nil, nil, err
	}

	var fullOrgsList []resources.Organization
	warnings, err := client.paginate(request, resources.Organization{}, func(item interface{}) error {
		if app, ok := item.(resources.Organization); ok {
			fullOrgsList = append(fullOrgsList, app)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   resources.Organization{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullOrgsList, warnings, err
}

// GetOrganization gets an organization by the given guid.
func (client *Client) GetOrganization(orgGUID string) (resources.Organization, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationRequest,
		URIParams:   map[string]string{"organization_guid": orgGUID},
	})

	if err != nil {
		return resources.Organization{}, nil, err
	}

	var responseOrg resources.Organization
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseOrg,
	}
	err = client.connection.Make(request, &response)
	if err != nil {
		return resources.Organization{}, response.Warnings, err
	}

	return responseOrg, response.Warnings, nil
}

// GetOrganizations lists organizations with optional filters.
func (client *Client) GetOrganizations(query ...Query) ([]resources.Organization, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationsRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullOrgsList []resources.Organization
	warnings, err := client.paginate(request, resources.Organization{}, func(item interface{}) error {
		if app, ok := item.(resources.Organization); ok {
			fullOrgsList = append(fullOrgsList, app)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   resources.Organization{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullOrgsList, warnings, err
}

// UpdateOrganization updates an organization with the given properties.
func (client *Client) UpdateOrganization(org resources.Organization) (resources.Organization, Warnings, error) {
	orgGUID := org.GUID
	org.GUID = ""
	orgBytes, err := json.Marshal(org)
	if err != nil {
		return resources.Organization{}, nil, err
	}
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchOrganizationRequest,
		Body:        bytes.NewReader(orgBytes),
		URIParams:   map[string]string{"organization_guid": orgGUID},
	})

	if err != nil {
		return resources.Organization{}, nil, err
	}

	var responseOrg resources.Organization
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseOrg,
	}
	err = client.connection.Make(request, &response)

	if err != nil {
		return resources.Organization{}, nil, err
	}
	return responseOrg, nil, err
}
