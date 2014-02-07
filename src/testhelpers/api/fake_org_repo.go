package api

import (
	"cf"
	"cf/models"
	"cf/net"
)

type FakeOrgRepository struct {
	Organizations []models.Organization

	CreateName      string
	CreateOrgExists bool

	FindByNameName         string
	FindByNameErr          bool
	FindByNameNotFound     bool
	FindByNameOrganization models.Organization

	RenameOrganizationGuid string
	RenameNewName          string

	DeletedOrganizationGuid string
}

func (repo FakeOrgRepository) ListOrgs(stop chan bool) (orgsChan chan []models.Organization, statusChan chan net.ApiResponse) {
	orgsChan = make(chan []models.Organization, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		orgsCount := len(repo.Organizations)
		for i := 0; i < orgsCount; i += 2 {
			select {
			case <-stop:
				break
			default:
				if orgsCount-i > 1 {
					orgsChan <- repo.Organizations[i : i+2]
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

func (repo *FakeOrgRepository) FindByName(name string) (org models.Organization, apiResponse net.ApiResponse) {
	repo.FindByNameName = name

	var foundOrg bool = false
	for _, anOrg := range repo.Organizations {
		if name == anOrg.Name {
			foundOrg = true
			org = anOrg
			break
		}
	}

	if repo.FindByNameErr || !foundOrg {
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
