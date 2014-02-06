package api

import (
	"cf"
	"cf/models"
	"cf/net"
)

type FakeSpaceRepository struct {
	CurrentSpace models.Space

	Spaces []models.Space

	FindByNameName     string
	FindByNameSpace    models.Space
	FindByNameErr      bool
	FindByNameNotFound bool

	FindByNameInOrgName    string
	FindByNameInOrgOrgGuid string
	FindByNameInOrgSpace   models.Space

	SummarySpace models.Space

	CreateSpaceName    string
	CreateSpaceOrgGuid string
	CreateSpaceExists  bool
	CreateSpaceSpace   models.Space

	RenameSpaceGuid string
	RenameNewName   string

	DeletedSpaceGuid string
}

func (repo FakeSpaceRepository) GetCurrentSpace() (space models.Space) {
	return repo.CurrentSpace
}

func (repo FakeSpaceRepository) ListSpaces(stop chan bool) (spacesChan chan []models.Space, statusChan chan net.ApiResponse) {
	spacesChan = make(chan []models.Space, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		spacesCount := len(repo.Spaces)
		for i := 0; i < spacesCount; i += 2 {
			select {
			case <-stop:
				break
			default:
				if spacesCount-i > 1 {
					spacesChan <- repo.Spaces[i : i+2]
				} else {
					spacesChan <- repo.Spaces[i:]
				}
			}
		}

		close(spacesChan)
		close(statusChan)

		cf.WaitForClose(stop)
	}()

	return
}

func (repo *FakeSpaceRepository) FindByName(name string) (space models.Space, apiResponse net.ApiResponse) {
	repo.FindByNameName = name

	var foundSpace bool = false
	for _, someSpace := range repo.Spaces {
		if name == someSpace.Name {
			foundSpace = true
			space = someSpace
			break
		}
	}

	if repo.FindByNameErr || !foundSpace {
		apiResponse = net.NewApiResponseWithMessage("Error finding space by name.")
	}

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Space", name)
	}

	return
}

func (repo *FakeSpaceRepository) FindByNameInOrg(name, orgGuid string) (space models.Space, apiResponse net.ApiResponse) {
	repo.FindByNameInOrgName = name
	repo.FindByNameInOrgOrgGuid = orgGuid
	space = repo.FindByNameInOrgSpace
	return
}

func (repo *FakeSpaceRepository) GetSummary() (space models.Space, apiResponse net.ApiResponse) {
	space = repo.SummarySpace
	return
}

func (repo *FakeSpaceRepository) Create(name string, orgGuid string) (space models.Space, apiResponse net.ApiResponse) {
	if repo.CreateSpaceExists {
		apiResponse = net.NewApiResponse("Space already exists", cf.SPACE_EXISTS, 400)
		return
	}
	repo.CreateSpaceName = name
	repo.CreateSpaceOrgGuid = orgGuid
	space = repo.CreateSpaceSpace
	return
}

func (repo *FakeSpaceRepository) Rename(spaceGuid, newName string) (apiResponse net.ApiResponse) {
	repo.RenameSpaceGuid = spaceGuid
	repo.RenameNewName = newName
	return
}

func (repo *FakeSpaceRepository) Delete(spaceGuid string) (apiResponse net.ApiResponse) {
	repo.DeletedSpaceGuid = spaceGuid
	return
}
