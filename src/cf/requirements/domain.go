package requirements

import (
	"cf"
	"cf/api"
	"cf/net"
	"cf/terminal"
)

type DomainRequirement interface {
	Requirement
	GetDomain() cf.Domain
}

type domainApiRequirement struct {
	name       string
	ui         terminal.UI
	domainRepo api.DomainRepository
	domain     cf.Domain
}

func NewDomainRequirement(name string, ui terminal.UI, domainRepo api.DomainRepository) (req *domainApiRequirement) {
	req = new(domainApiRequirement)
	req.name = name
	req.ui = ui
	req.domainRepo = domainRepo
	return
}

func (req *domainApiRequirement) Execute() bool {
	var apiResponse net.ApiResponse
	req.domain, apiResponse = req.domainRepo.FindByNameInCurrentSpace(req.name)

	if apiResponse.IsNotSuccessful() {
		req.ui.Failed(apiResponse.Message)
		return false
	}

	return true
}

func (req *domainApiRequirement) GetDomain() cf.Domain {
	return req.domain
}
