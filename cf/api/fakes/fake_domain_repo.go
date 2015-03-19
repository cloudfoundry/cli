package fakes

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
)

type FakeDomainRepository struct {
	ListDomainsForOrgGuid        string
	ListDomainsForOrgDomains     []models.DomainFields
	ListDomainsForOrgApiResponse error

	FindByNameInOrgName        string
	FindByNameInOrgGuid        string
	FindByNameInOrgDomain      []models.DomainFields
	FindByNameInOrgApiResponse error

	FindSharedByNameName     string
	FindSharedByNameDomain   models.DomainFields
	FindSharedByNameNotFound bool
	FindSharedByNameErr      bool

	FindPrivateByNameName     string
	FindPrivateByNameDomain   models.DomainFields
	FindPrivateByNameNotFound bool
	FindPrivateByNameErr      bool

	CreateDomainName          string
	CreateDomainOwningOrgGuid string

	CreateSharedDomainName string

	DeleteDomainGuid  string
	DeleteApiResponse error

	DeleteSharedDomainGuid  string
	DeleteSharedApiResponse error

	domainCursor int
}

func (repo *FakeDomainRepository) ListDomainsForOrg(orgGuid string, cb func(models.DomainFields) bool) error {
	repo.ListDomainsForOrgGuid = orgGuid
	for _, d := range repo.ListDomainsForOrgDomains {
		cb(d)
	}
	return repo.ListDomainsForOrgApiResponse
}

func (repo *FakeDomainRepository) FindSharedByName(name string) (domain models.DomainFields, apiErr error) {
	repo.FindSharedByNameName = name
	domain = repo.FindSharedByNameDomain

	if repo.FindSharedByNameNotFound {
		apiErr = errors.NewModelNotFoundError("Domain", name)
	}
	if repo.FindSharedByNameErr {
		apiErr = errors.New("Error finding domain")
	}

	return
}

func (repo *FakeDomainRepository) FindPrivateByName(name string) (domain models.DomainFields, apiErr error) {
	repo.FindPrivateByNameName = name
	domain = repo.FindPrivateByNameDomain

	if repo.FindPrivateByNameNotFound {
		apiErr = errors.NewModelNotFoundError("Domain", name)
	}
	if repo.FindPrivateByNameErr {
		apiErr = errors.New("Error finding domain")
	}

	return
}

func (repo *FakeDomainRepository) FindByNameInOrg(name string, owningOrgGuid string) (domain models.DomainFields, apiErr error) {
	repo.FindByNameInOrgName = name
	repo.FindByNameInOrgGuid = owningOrgGuid
	if len(repo.FindByNameInOrgDomain) == 0 {
		domain = models.DomainFields{}
	} else {
		domain = repo.FindByNameInOrgDomain[repo.domainCursor]
		repo.domainCursor++
	}
	apiErr = repo.FindByNameInOrgApiResponse
	return
}

func (repo *FakeDomainRepository) Create(domainName string, owningOrgGuid string) (createdDomain models.DomainFields, apiErr error) {
	repo.CreateDomainName = domainName
	repo.CreateDomainOwningOrgGuid = owningOrgGuid
	return
}

func (repo *FakeDomainRepository) CreateSharedDomain(domainName string) (apiErr error) {
	repo.CreateSharedDomainName = domainName
	return
}

func (repo *FakeDomainRepository) Delete(domainGuid string) (apiErr error) {
	repo.DeleteDomainGuid = domainGuid
	apiErr = repo.DeleteApiResponse
	return
}

func (repo *FakeDomainRepository) DeleteSharedDomain(domainGuid string) (apiErr error) {
	repo.DeleteSharedDomainGuid = domainGuid
	apiErr = repo.DeleteSharedApiResponse
	return
}

func (repo *FakeDomainRepository) FirstOrDefault(orgGuid string, name *string) (domain models.DomainFields, error error) {
	if name == nil {
		domain, error = repo.defaultDomain(orgGuid)
	} else {
		domain, error = repo.FindByNameInOrg(*name, orgGuid)
	}
	return
}

func (repo *FakeDomainRepository) defaultDomain(orgGuid string) (models.DomainFields, error) {
	var foundDomain *models.DomainFields
	repo.ListDomainsForOrg(orgGuid, func(domain models.DomainFields) bool {
		foundDomain = &domain
		return !domain.Shared
	})

	if foundDomain == nil {
		return models.DomainFields{}, errors.New("Could not find a default domain")
	}

	return *foundDomain, nil
}
