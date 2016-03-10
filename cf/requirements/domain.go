package requirements

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
)

//go:generate counterfeiter -o fakes/fake_domain_requirement.go . DomainRequirement
type DomainRequirement interface {
	Requirement
	GetDomain() models.DomainFields
}

type domainApiRequirement struct {
	name       string
	config     core_config.Reader
	domainRepo api.DomainRepository
	domain     models.DomainFields
}

func NewDomainRequirement(name string, config core_config.Reader, domainRepo api.DomainRepository) (req *domainApiRequirement) {
	req = new(domainApiRequirement)
	req.name = name
	req.config = config
	req.domainRepo = domainRepo
	return
}

func (req *domainApiRequirement) Execute() error {
	var apiErr error
	req.domain, apiErr = req.domainRepo.FindByNameInOrg(req.name, req.config.OrganizationFields().Guid)

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (req *domainApiRequirement) GetDomain() models.DomainFields {
	return req.domain
}
