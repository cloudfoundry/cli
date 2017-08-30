package v2action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	log "github.com/sirupsen/logrus"
)

// Domain represents a CLI Domain.
type Domain ccv2.Domain

// DomainNotFoundError is an error wrapper that represents the case
// when the domain is not found.
type DomainNotFoundError struct{}

// Error method to display the error message.
func (DomainNotFoundError) Error() string {
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

func (actor Actor) GetDomainsByNameAndOrganization(domainNames []string, orgGUID string) ([]Domain, Warnings, error) {
	var domains []Domain
	var allWarnings Warnings

	// TODO: If the following causes URI length problems, break domainNames into
	// batched (based on character length?) and loop over them.

	sharedDomains, warnings, err := actor.CloudControllerClient.GetSharedDomains(ccv2.Query{
		Filter:   ccv2.NameFilter,
		Operator: ccv2.InOperator,
		Values:   domainNames,
	})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	for _, domain := range sharedDomains {
		domains = append(domains, Domain(domain))
		actor.saveDomain(domain)
	}

	privateDomains, warnings, err := actor.CloudControllerClient.GetOrganizationPrivateDomains(
		orgGUID,
		ccv2.Query{
			Filter:   ccv2.NameFilter,
			Operator: ccv2.InOperator,
			Values:   domainNames,
		})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	for _, domain := range privateDomains {
		domains = append(domains, Domain(domain))
		actor.saveDomain(domain)
	}

	return domains, allWarnings, err
}

// GetSharedDomain returns the shared domain associated with the provided
// Domain GUID.
func (actor Actor) GetSharedDomain(domainGUID string) (Domain, Warnings, error) {
	if domain, found := actor.loadDomain(domainGUID); found {
		log.WithFields(log.Fields{
			"domain": domain.Name,
			"GUID":   domain.GUID,
		}).Debug("using domain from cache")
		return domain, nil, nil
	}

	domain, warnings, err := actor.CloudControllerClient.GetSharedDomain(domainGUID)
	if isResourceNotFoundError(err) {
		return Domain{}, Warnings(warnings), DomainNotFoundError{}
	}

	actor.saveDomain(domain)
	return Domain(domain), Warnings(warnings), err
}

// GetPrivateDomain returns the private domain associated with the provided
// Domain GUID.
func (actor Actor) GetPrivateDomain(domainGUID string) (Domain, Warnings, error) {
	if domain, found := actor.loadDomain(domainGUID); found {
		log.WithFields(log.Fields{
			"domain": domain.Name,
			"GUID":   domain.GUID,
		}).Debug("using domain from cache")
		return domain, nil, nil
	}

	domain, warnings, err := actor.CloudControllerClient.GetPrivateDomain(domainGUID)
	if isResourceNotFoundError(err) {
		return Domain{}, Warnings(warnings), DomainNotFoundError{}
	}

	actor.saveDomain(domain)
	return Domain(domain), Warnings(warnings), err
}

// GetOrganizationDomains returns the shared and private domains associated
// with an organization.
func (actor Actor) GetOrganizationDomains(orgGUID string) ([]Domain, Warnings, error) {
	var allWarnings Warnings
	var allDomains []Domain

	domains, warnings, err := actor.CloudControllerClient.GetSharedDomains()
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return []Domain{}, allWarnings, err
	}
	for _, domain := range domains {
		allDomains = append(allDomains, Domain(domain))
	}

	domains, warnings, err = actor.CloudControllerClient.GetOrganizationPrivateDomains(orgGUID)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return []Domain{}, allWarnings, err
	}
	for _, domain := range domains {
		allDomains = append(allDomains, Domain(domain))
	}

	return allDomains, allWarnings, nil
}

func (actor Actor) saveDomain(domain ccv2.Domain) {
	if domain.GUID != "" {
		actor.domainCache[domain.GUID] = Domain(domain)
	}
}

func (actor Actor) loadDomain(domainGUID string) (Domain, bool) {
	domain, found := actor.domainCache[domainGUID]
	return domain, found
}
