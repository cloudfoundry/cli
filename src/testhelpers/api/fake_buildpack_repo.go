package api

import (
	"cf"
	"cf/errors"
	"cf/models"
)

type FakeBuildpackRepository struct {
	Buildpacks []models.Buildpack

	FindByNameNotFound    bool
	FindByNameName        string
	FindByNameBuildpack   models.Buildpack
	FindByNameApiResponse errors.Error

	CreateBuildpackExists bool
	CreateBuildpack       models.Buildpack
	CreateApiResponse     errors.Error

	DeleteBuildpackGuid string
	DeleteApiResponse   errors.Error

	UpdateBuildpack models.Buildpack
}

func (repo *FakeBuildpackRepository) ListBuildpacks(cb func(models.Buildpack) bool) errors.Error {
	for _, b := range repo.Buildpacks {
		cb(b)
	}
	return nil
}

func (repo *FakeBuildpackRepository) FindByName(name string) (buildpack models.Buildpack, apiResponse errors.Error) {
	repo.FindByNameName = name
	buildpack = repo.FindByNameBuildpack

	if repo.FindByNameNotFound {
		apiResponse = errors.NewNotFoundError("Buildpack %s not found", name)
	}

	return
}

func (repo *FakeBuildpackRepository) Create(name string, position *int, enabled *bool, locked *bool) (createdBuildpack models.Buildpack, apiResponse errors.Error) {
	if repo.CreateBuildpackExists {
		return repo.CreateBuildpack, errors.NewError("Buildpack already exists", cf.BUILDPACK_EXISTS)
	}

	repo.CreateBuildpack = models.Buildpack{Name: name, Position: position, Enabled: enabled, Locked: locked}
	return repo.CreateBuildpack, repo.CreateApiResponse
}

func (repo *FakeBuildpackRepository) Delete(buildpackGuid string) (apiResponse errors.Error) {
	repo.DeleteBuildpackGuid = buildpackGuid
	apiResponse = repo.DeleteApiResponse
	return
}

func (repo *FakeBuildpackRepository) Update(buildpack models.Buildpack) (updatedBuildpack models.Buildpack, apiResponse errors.Error) {
	repo.UpdateBuildpack = buildpack
	return
}
