package api

import (
	"cf"
	"cf/net"
)

type FakeDomainRepository struct {
	DefaultAppDomain cf.Domain

	FindAllInCurrentSpaceDomains []cf.Domain

	ListDomainsForOrgDomainsGuid string
	ListDomainsForOrgDomains []cf.Domain

	FindByNameInOrgDomain cf.Domain
	FindByNameInOrgApiResponse net.ApiResponse

	FindByNameInCurrentSpaceName string

	FindByNameName string
	FindByNameDomain cf.Domain
	FindByNameNotFound bool
	FindByNameErr bool

	CreateDomainName string
	CreateDomainOwningOrgGuid string

	CreateSharedDomainName string

	DeleteDomainGuid string
	DeleteApiResponse net.ApiResponse
}

func (repo *FakeDomainRepository) FindDefaultAppDomain() (domain cf.Domain, apiResponse net.ApiResponse){
	domain = repo.DefaultAppDomain
	return
}

func (repo *FakeDomainRepository) ListDomainsForOrg(orgGuid string, stop chan bool) (domainsChan chan []cf.Domain, statusChan chan net.ApiResponse){
	repo.ListDomainsForOrgDomainsGuid = orgGuid

	domainsChan = make(chan []cf.Domain, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		domainsCount := len(repo.ListDomainsForOrgDomains)
		for i:= 0; i < domainsCount; i += 2 {
			select {
			case <-stop:
				break
			default:
				if domainsCount - i > 1 {
					domainsChan <- repo.ListDomainsForOrgDomains[i:i+2]
				} else {
					domainsChan <- repo.ListDomainsForOrgDomains[i:]
				}
			}
		}

		close(domainsChan)
		close(statusChan)

		cf.WaitForClose(stop)
	}()

	return
}

func (repo *FakeDomainRepository) FindByName(name string) (domain cf.Domain, apiResponse net.ApiResponse){
	repo.FindByNameName = name
	domain = repo.FindByNameDomain

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found","Domain", name)
	}
	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding domain")
	}

	return
}


func (repo *FakeDomainRepository) FindByNameInCurrentSpace(name string) (domain cf.Domain, apiResponse net.ApiResponse){
	repo.FindByNameInCurrentSpaceName = name
	domain = repo.FindByNameDomain

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found","Domain", name)
	}
	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding domain")
	}

	return
}

func (repo *FakeDomainRepository) FindByNameInOrg(name string, owningOrgGuid string) (domain cf.Domain, apiResponse net.ApiResponse) {
	domain = repo.FindByNameInOrgDomain
	apiResponse = repo.FindByNameInOrgApiResponse
	return
}

func (repo *FakeDomainRepository) Create(domainName string, owningOrgGuid string) (createdDomain cf.DomainFields, apiResponse net.ApiResponse){
	repo.CreateDomainName = domainName
	repo.CreateDomainOwningOrgGuid = owningOrgGuid
	return
}

func (repo *FakeDomainRepository) CreateSharedDomain(domainName string) (apiResponse net.ApiResponse){
	repo.CreateSharedDomainName = domainName
	return
}

func (repo *FakeDomainRepository) Delete(domainGuid string) (apiResponse net.ApiResponse) {
	repo.DeleteDomainGuid = domainGuid
	apiResponse = repo.DeleteApiResponse
	return
}
