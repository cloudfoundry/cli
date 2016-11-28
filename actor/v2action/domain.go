package v2action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
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

func isResourceNotFoundError(err error) bool {
	_, isResourceNotFound := err.(cloudcontroller.ResourceNotFoundError)
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
