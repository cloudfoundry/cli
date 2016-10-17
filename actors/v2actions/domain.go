package v2actions

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

// Domain represents a CLI Domain.
type Domain ccv2.Domain

// DomainNotFoundError is an error wrapper that represents the case
// when the domain is not found.
type DomainNotFoundError struct{}

// Error method to display the error message.
func (e DomainNotFoundError) Error() string {
	return "Domain not found."
}

func isResourceNotFoundError(err error) bool {
	_, isResourceNotFound := err.(ccv2.ResourceNotFoundError)

	return isResourceNotFound
}

// GetDomain returns a shared or private domain with the domain GUID.
func (actor Actor) GetDomain(domainGUID string) (Domain, Warnings, error) {
	var allWarnings Warnings

	domain, warnings, err := actor.CloudControllerClient.GetSharedDomain(domainGUID)
	allWarnings = append(allWarnings, warnings...)
	if err == nil {
		return Domain(domain), allWarnings, nil
	}

	if !isResourceNotFoundError(err) {
		return Domain{}, allWarnings, err
	}

	domain, warnings, err = actor.CloudControllerClient.GetPrivateDomain(domainGUID)
	allWarnings = append(allWarnings, warnings...)
	if err == nil {
		return Domain(domain), allWarnings, nil
	}

	if isResourceNotFoundError(err) {
		return Domain{}, allWarnings, DomainNotFoundError{}
	}

	return Domain{}, allWarnings, err
}
