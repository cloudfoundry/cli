package pushaction

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	log "github.com/sirupsen/logrus"
)

// DefaultDomain looks up the shared and then private domains and returns back
// the first one in the list as the default.
func (actor Actor) DefaultDomain(orgGUID string) (v2action.Domain, Warnings, error) {
	log.Infoln("getting org domains for org GUID:", orgGUID)
	// the domains object contains all the shared domains AND all domains private to this org
	domains, warnings, err := actor.V2Actor.GetOrganizationDomains(orgGUID)
	if err != nil {
		log.Errorln("searching for domains in org:", err)
		return v2action.Domain{}, Warnings(warnings), err
	}

	if len(domains) == 0 {
		log.Error("no domains found")
		return v2action.Domain{}, Warnings(warnings), actionerror.NoDomainsFoundError{OrganizationGUID: orgGUID}
	}

	log.Debugf("selecting first domain as default domain: %#v", domains)
	return domains[0], Warnings(warnings), nil
}
