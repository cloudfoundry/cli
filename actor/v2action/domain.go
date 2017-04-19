package v2action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// Domain represents a CLI Domain.
type Domain ccv2.Domain

// DomainNotFoundError is an error wrapper that represents the case
// when the domain is not found.
type DomainNotFoundError struct{}

// Error method to display the error message.
func (e DomainNotFoundError) Error() string {
	return "Domain not found."
}

// TODO: Move into own file or add function to CCV2/3
func isResourceNotFoundError(err error) bool {
	_, isResourceNotFound := err.(ccerror.ResourceNotFoundError)
	return isResourceNotFound
}

// GetDomain returns the shared or private domain associated with the provided
// Domain GUID.
func (actor Actor) GetDomain(domainGUID string) (Domain, Warnings, error) {
	var allWarnings Warnings

	domain, warnings, err := actor.GetSharedDomain(domainGUID)
	allWarnings = append(allWarnings, warnings...)
	switch err.(type) {
	case nil:
		return domain, allWarnings, nil
	case DomainNotFoundError:
	default:
		return Domain{}, allWarnings, err
	}

	domain, warnings, err = actor.GetPrivateDomain(domainGUID)
	allWarnings = append(allWarnings, warnings...)
	switch err.(type) {
	case nil:
		return domain, allWarnings, nil
	default:
		return Domain{}, allWarnings, err
	}
}

// GetSharedDomain returns the shared domain associated with the provided
// Domain GUID.
func (actor Actor) GetSharedDomain(domainGUID string) (Domain, Warnings, error) {
	domain, warnings, err := actor.CloudControllerClient.GetSharedDomain(domainGUID)
	if err == nil {
		return Domain(domain), Warnings(warnings), nil
	}

	if isResourceNotFoundError(err) {
		return Domain{}, Warnings(warnings), DomainNotFoundError{}
	}

	return Domain{}, Warnings(warnings), err
}

// GetPrivateDomain returns the private domain associated with the provided
// Domain GUID.
func (actor Actor) GetPrivateDomain(domainGUID string) (Domain, Warnings, error) {
	domain, warnings, err := actor.CloudControllerClient.GetPrivateDomain(domainGUID)
	if err == nil {
		return Domain(domain), Warnings(warnings), nil
	}

	if isResourceNotFoundError(err) {
		return Domain{}, Warnings(warnings), DomainNotFoundError{}
	}

	return Domain{}, Warnings(warnings), err
}

// GetOrganizationDomains returns the private and shared domains associated
// with an organization. Private domains will be listed before shared.
func (actor Actor) GetOrganizationDomains(orgGUID string) ([]Domain, Warnings, error) {
	var allWarnings Warnings
	var allDomains []Domain

	// push requires that private domains are listed first to deterime the
	// default domain.
	domains, warnings, err := actor.CloudControllerClient.GetOrganizationPrivateDomains(orgGUID, nil)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return []Domain{}, allWarnings, err
	}
	for _, domain := range domains {
		allDomains = append(allDomains, Domain(domain))
	}

	domains, warnings, err = actor.CloudControllerClient.GetSharedDomains()
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return []Domain{}, allWarnings, err
	}
	for _, domain := range domains {
		allDomains = append(allDomains, Domain(domain))
	}

	return allDomains, allWarnings, nil
}
