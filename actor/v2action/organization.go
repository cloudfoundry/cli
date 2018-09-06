package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	uaaconst "code.cloudfoundry.org/cli/api/uaa/constant"
)

// Organization represents a CLI Organization.
type Organization ccv2.Organization

// GetOrganization returns an Organization based on the provided guid.
func (actor Actor) GetOrganization(guid string) (Organization, Warnings, error) {
	org, warnings, err := actor.CloudControllerClient.GetOrganization(guid)

	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return Organization{}, Warnings(warnings), actionerror.OrganizationNotFoundError{GUID: guid}
	}

	return Organization(org), Warnings(warnings), err
}

// GetOrganizationByName returns an Organization based off of the name given.
func (actor Actor) GetOrganizationByName(orgName string) (Organization, Warnings, error) {
	orgs, warnings, err := actor.CloudControllerClient.GetOrganizations(ccv2.Filter{
		Type:     constant.NameFilter,
		Operator: constant.EqualOperator,
		Values:   []string{orgName},
	})
	if err != nil {
		return Organization{}, Warnings(warnings), err
	}

	if len(orgs) == 0 {
		return Organization{}, Warnings(warnings), actionerror.OrganizationNotFoundError{Name: orgName}
	}

	if len(orgs) > 1 {
		var guids []string
		for _, org := range orgs {
			guids = append(guids, org.GUID)
		}
		return Organization{}, Warnings(warnings), actionerror.MultipleOrganizationsFoundError{Name: orgName, GUIDs: guids}
	}

	return Organization(orgs[0]), Warnings(warnings), nil
}

// GrantOrgManagerByUsername gives the Org Manager role to the provided user.
func (actor Actor) GrantOrgManagerByUsername(guid string, username string) (Warnings, error) {
	var warnings ccv2.Warnings
	var err error

	if actor.Config.UAAGrantType() != string(uaaconst.GrantTypeClientCredentials) {
		warnings, err = actor.CloudControllerClient.UpdateOrganizationManagerByUsername(guid, username)
	} else {
		warnings, err = actor.CloudControllerClient.UpdateOrganizationManager(guid, username)
	}

	return Warnings(warnings), err
}

// CreateOrganization creates an Organization based on the provided orgName.
func (actor Actor) CreateOrganization(orgName string, quotaName string) (Organization, Warnings, error) {
	var quotaGUID string
	var allWarnings Warnings

	if quotaName != "" {
		quota, warnings, err := actor.GetOrganizationQuotaByName(quotaName)

		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return Organization{}, allWarnings, err
		}

		quotaGUID = quota.GUID
	}

	org, warnings, err := actor.CloudControllerClient.CreateOrganization(orgName, quotaGUID)
	allWarnings = append(allWarnings, warnings...)
	if _, ok := err.(ccerror.OrganizationNameTakenError); ok {
		return Organization{}, allWarnings, actionerror.OrganizationNameTakenError{Name: orgName}
	}
	if err != nil {
		return Organization{}, allWarnings, err
	}

	return Organization(org), allWarnings, nil
}

// DeleteOrganization deletes the Organization associated with the provided
// GUID. Once the deletion request is sent, it polls the deletion job until
// it's finished.
func (actor Actor) DeleteOrganization(orgName string) (Warnings, error) {
	var allWarnings Warnings

	org, warnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	job, deleteWarnings, err := actor.CloudControllerClient.DeleteOrganizationJob(org.GUID)
	allWarnings = append(allWarnings, deleteWarnings...)
	if err != nil {
		return allWarnings, err
	}

	ccWarnings, err := actor.CloudControllerClient.PollJob(job)
	for _, warning := range ccWarnings {
		allWarnings = append(allWarnings, warning)
	}

	return allWarnings, err
}

func (actor Actor) GetOrganizations() ([]Organization, Warnings, error) {
	var returnedOrgs []Organization
	orgs, warnings, err := actor.CloudControllerClient.GetOrganizations()
	for _, org := range orgs {
		returnedOrgs = append(returnedOrgs, Organization(org))
	}
	return returnedOrgs, Warnings(warnings), err
}
