package v2actions

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

type Domain ccv2.Domain

type DomainNotFoundError struct{}

func (e DomainNotFoundError) Error() string {
	return "Domain not found."
}

func (actor Actor) GetDomainByGUID(domainGUID string) (Domain, Warnings, error) {
	var allWarnings Warnings

	domain, warnings, err := actor.CloudControllerClient.GetSharedDomain(domainGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return Domain{}, allWarnings, err
	}

	if domain.GUID != "" {
		return Domain(domain), allWarnings, nil
	}

	domain, warnings, err = actor.CloudControllerClient.GetPrivateDomain(domainGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return Domain{}, allWarnings, err
	}

	if domain.GUID == "" {
		return Domain{}, allWarnings, DomainNotFoundError{}
	}

	return Domain(domain), allWarnings, nil
}
