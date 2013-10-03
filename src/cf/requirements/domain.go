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

type DomainApiRequirement struct {
	name       string
	ui         terminal.UI
	domainRepo api.DomainRepository
	domain     cf.Domain
}

func NewDomainRequirement(name string, ui terminal.UI, domainRepo api.DomainRepository) (req *DomainApiRequirement) {
	req = new(DomainApiRequirement)
	req.name = name
	req.ui = ui
	req.domainRepo = domainRepo
	return
}

func (req *DomainApiRequirement) Execute() bool {
	var apiStatus net.ApiStatus
	req.domain, apiStatus = req.domainRepo.FindByNameInCurrentSpace(req.name)

	if apiStatus.NotSuccessful() {
		req.ui.Failed(apiStatus.Message)
		return false
	}

	return true
}

func (req *DomainApiRequirement) GetDomain() cf.Domain {
	return req.domain
}
