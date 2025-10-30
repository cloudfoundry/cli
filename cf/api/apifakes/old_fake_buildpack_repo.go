package apifakes

import (
	"code.cloudfoundry.org/cli/v8/cf/errors"
	"code.cloudfoundry.org/cli/v8/cf/models"
)

type OldFakeBuildpackRepository struct {
	Buildpacks []models.Buildpack

	FindByNameNotFound    bool
	FindByNameName        string
	FindByNameBuildpack   models.Buildpack
	FindByNameAPIResponse error
	FindByNameAmbiguous   bool

	FindByNameAndStackNotFound  bool
	FindByNameAndStackName      string
	FindByNameAndStackStack     string
	FindByNameAndStackBuildpack models.Buildpack

	FindByNameWithNilStackNotFound  bool
	FindByNameWithNilStackName      string
	FindByNameWithNilStackBuildpack models.Buildpack

	CreateBuildpackExists             bool
	CreateBuildpackWithNilStackExists bool
	CreateBuildpack                   models.Buildpack
	CreateAPIResponse                 error

	DeleteBuildpackGUID string
	DeleteAPIResponse   error

	UpdateBuildpackArgs struct {
		Buildpack models.Buildpack
	}

	UpdateBuildpackReturns struct {
		Error error
	}
}

func (repo *OldFakeBuildpackRepository) ListBuildpacks(cb func(models.Buildpack) bool) error {
	for _, b := range repo.Buildpacks {
		cb(b)
	}
	return nil
}

func (repo *OldFakeBuildpackRepository) FindByName(name string) (buildpack models.Buildpack, apiErr error) {
	repo.FindByNameName = name
	buildpack = repo.FindByNameBuildpack

	if repo.FindByNameNotFound {
		apiErr = errors.NewModelNotFoundError("Buildpack", name)
	} else if repo.FindByNameAmbiguous {
		apiErr = errors.NewAmbiguousModelError("Buildpack", name)
	}

	return
}

func (repo *OldFakeBuildpackRepository) FindByNameAndStack(name, stack string) (buildpack models.Buildpack, apiErr error) {
	repo.FindByNameAndStackName = name
	repo.FindByNameAndStackStack = stack
	buildpack = repo.FindByNameAndStackBuildpack

	if repo.FindByNameAndStackNotFound {
		apiErr = errors.NewModelNotFoundError("Buildpack", name)
	}

	return
}

func (repo *OldFakeBuildpackRepository) FindByNameWithNilStack(name string) (buildpack models.Buildpack, apiErr error) {
	repo.FindByNameWithNilStackName = name
	buildpack = repo.FindByNameWithNilStackBuildpack

	if repo.FindByNameWithNilStackNotFound {
		apiErr = errors.NewModelNotFoundError("Buildpack", name)
	}

	return
}

func (repo *OldFakeBuildpackRepository) Create(name string, position *int, enabled *bool, locked *bool) (createdBuildpack models.Buildpack, apiErr error) {
	if repo.CreateBuildpackExists {
		return repo.CreateBuildpack, errors.NewHTTPError(400, errors.BuildpackNameTaken, "Buildpack already exists")
	} else if repo.CreateBuildpackWithNilStackExists {
		return repo.CreateBuildpack, errors.NewHTTPError(400, errors.StackUnique, "stack unique")
	}

	repo.CreateBuildpack = models.Buildpack{Name: name, Position: position, Enabled: enabled, Locked: locked, GUID: "BUILDPACK-GUID-X"}
	return repo.CreateBuildpack, repo.CreateAPIResponse
}

func (repo *OldFakeBuildpackRepository) Delete(buildpackGUID string) (apiErr error) {
	repo.DeleteBuildpackGUID = buildpackGUID
	apiErr = repo.DeleteAPIResponse
	return
}

func (repo *OldFakeBuildpackRepository) Update(buildpack models.Buildpack) (updatedBuildpack models.Buildpack, apiErr error) {
	repo.UpdateBuildpackArgs.Buildpack = buildpack
	apiErr = repo.UpdateBuildpackReturns.Error
	return
}
