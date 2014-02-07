package requirements

import (
	"cf/api"
	"cf/configuration"
	"cf/models"
	"cf/net"
	"cf/terminal"
)

type DomainRequirement interface {
	Requirement
	GetDomain() models.DomainFields
}

type domainApiRequirement struct {
	name       string
	ui         terminal.UI
	config     *configuration.Configuration
	domainRepo api.DomainRepository
	domain     models.DomainFields
}

func NewDomainRequirement(name string, ui terminal.UI, config *configuration.Configuration, domainRepo api.DomainRepository) (req *domainApiRequirement) {
	req = new(domainApiRequirement)
	req.name = name
	req.ui = ui
	req.config = config
	req.domainRepo = domainRepo
	return
}

func (req *domainApiRequirement) Execute() bool {
	var apiResponse net.ApiResponse
	req.domain, apiResponse = req.domainRepo.FindByNameInOrg(req.name, req.config.OrganizationFields.Guid)

	if apiResponse.IsNotSuccessful() {
		req.ui.Failed(apiResponse.Message)
		return false
	}

	return true
}

func (req *domainApiRequirement) GetDomain() models.DomainFields {
	return req.domain
}
