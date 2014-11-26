package fakes

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
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
	FindByNameInOrgError   error

	SummarySpace models.Space

	CreateSpaceName           string
	CreateSpaceOrgGuid        string
	CreateSpaceSpaceQuotaGuid string
	CreateSpaceExists         bool
	CreateSpaceSpace          models.Space

	RenameSpaceGuid string
	RenameNewName   string

	DeletedSpaceGuid string
}

func (repo FakeSpaceRepository) GetCurrentSpace() (space models.Space) {
	return repo.CurrentSpace
}

func (repo FakeSpaceRepository) ListSpaces(callback func(models.Space) bool) error {
	for _, space := range repo.Spaces {
		if !callback(space) {
			break
		}
	}
	return nil
}

func (repo *FakeSpaceRepository) FindByName(name string) (space models.Space, apiErr error) {
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
		apiErr = errors.New("Error finding space by name.")
	}

	if repo.FindByNameNotFound {
		apiErr = errors.NewModelNotFoundError("Space", name)
	}

	return
}

func (repo *FakeSpaceRepository) FindByNameInOrg(name, orgGuid string) (space models.Space, apiErr error) {
	repo.FindByNameInOrgName = name
	repo.FindByNameInOrgOrgGuid = orgGuid
	space = repo.FindByNameInOrgSpace
	apiErr = repo.FindByNameInOrgError
	return
}

func (repo *FakeSpaceRepository) GetSummary() (space models.Space, apiErr error) {
	space = repo.SummarySpace
	return
}

func (repo *FakeSpaceRepository) Create(name, orgGuid, spaceQuotaGuid string) (space models.Space, apiErr error) {
	if repo.CreateSpaceExists {
		apiErr = errors.NewHttpError(400, errors.SPACE_EXISTS, "Space already exists")
		return
	}
	repo.CreateSpaceName = name
	repo.CreateSpaceOrgGuid = orgGuid
	repo.CreateSpaceSpaceQuotaGuid = spaceQuotaGuid
	space = repo.CreateSpaceSpace
	return
}

func (repo *FakeSpaceRepository) Rename(spaceGuid, newName string) (apiErr error) {
	repo.RenameSpaceGuid = spaceGuid
	repo.RenameNewName = newName
	return
}

func (repo *FakeSpaceRepository) Delete(spaceGuid string) (apiErr error) {
	repo.DeletedSpaceGuid = spaceGuid
	return
}
