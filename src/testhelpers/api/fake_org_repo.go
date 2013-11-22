package api

import (
	"cf"
	"cf/net"
)

type FakeOrgRepository struct {
	Organizations []cf.Organization

	CreateName      string
	CreateOrgExists bool

	FindByNameName         string
	FindByNameErr          bool
	FindByNameNotFound     bool
	FindByNameOrganization cf.Organization

	RenameOrganizationGuid string
	RenameNewName      string

	DeletedOrganizationGuid string
}

func (repo FakeOrgRepository) ListOrgs(stop chan bool) (orgsChan chan []cf.Organization, statusChan chan net.ApiResponse) {
	orgsChan = make(chan []cf.Organization, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		orgsCount := len(repo.Organizations)
		for i:= 0; i < orgsCount; i += 2 {
			select {
			case <-stop:
				break
			default:
				if orgsCount - i > 1 {
					orgsChan <- repo.Organizations[i:i+2]
				} else {
					orgsChan <- repo.Organizations[i:]
				}
			}
		}

		close(orgsChan)
		close(statusChan)

		cf.WaitForClose(stop)
	}()

	return
}

func (repo *FakeOrgRepository) FindByName(name string) (org cf.Organization, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	org = repo.FindByNameOrganization

	if repo.FindByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding organization by name.")
	}

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Org", name)
	}

	return
}

func (repo *FakeOrgRepository) Create(name string) (apiResponse net.ApiResponse) {
	if repo.CreateOrgExists {
		apiResponse = net.NewApiResponse("Space already exists", cf.ORG_EXISTS, 400)
		return
	}
	repo.CreateName = name
	return
}

func (repo *FakeOrgRepository) Rename(orgGuid string, name string) (apiResponse net.ApiResponse) {
	repo.RenameOrganizationGuid = orgGuid
	repo.RenameNewName = name
	return
}

func (repo *FakeOrgRepository) Delete(orgGuid string) (apiResponse net.ApiResponse) {
	repo.DeletedOrganizationGuid = orgGuid
	return
}
