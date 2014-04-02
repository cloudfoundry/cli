package api

import (
	"cf/errors"
	"cf/models"
)

type FakeBuildpackRepository struct {
	Buildpacks []models.Buildpack

	FindByNameNotFound    bool
	FindByNameName        string
	FindByNameBuildpack   models.Buildpack
	FindByNameApiResponse error

	CreateBuildpackExists bool
	CreateBuildpack       models.Buildpack
	CreateApiResponse     error

	DeleteBuildpackGuid string
	DeleteApiResponse   error

	UpdateBuildpack   models.Buildpack
	RenameBuildpack   models.Buildpack
	RenameApiResponse error
}

func (repo *FakeBuildpackRepository) ListBuildpacks(cb func(models.Buildpack) bool) error {
	for _, b := range repo.Buildpacks {
		cb(b)
	}
	return nil
}

func (repo *FakeBuildpackRepository) FindByName(name string) (buildpack models.Buildpack, apiErr error) {
	repo.FindByNameName = name
	buildpack = repo.FindByNameBuildpack

	if repo.FindByNameNotFound {
		apiErr = errors.NewModelNotFoundError("Buildpack", name)
	}

	return
}

func (repo *FakeBuildpackRepository) Create(name string, position *int, enabled *bool, locked *bool) (createdBuildpack models.Buildpack, apiErr error) {
	if repo.CreateBuildpackExists {
		return repo.CreateBuildpack, errors.NewHttpError(400, errors.BUILDPACK_EXISTS, "Buildpack already exists")
	}

	repo.CreateBuildpack = models.Buildpack{Name: name, Position: position, Enabled: enabled, Locked: locked}
	return repo.CreateBuildpack, repo.CreateApiResponse
}

func (repo *FakeBuildpackRepository) Delete(buildpackGuid string) (apiErr error) {
	repo.DeleteBuildpackGuid = buildpackGuid
	apiErr = repo.DeleteApiResponse
	return
}

func (repo *FakeBuildpackRepository) Update(buildpack models.Buildpack) (updatedBuildpack models.Buildpack, apiErr error) {
	repo.UpdateBuildpack = buildpack
	return
}

func (repo *FakeBuildpackRepository) Rename(buildpack models.Buildpack, newbuildpackName string) (updatedBuildpack models.Buildpack, apiErr error) {
	buildpack.Name = newbuildpackName
	repo.RenameBuildpack = buildpack
	apiErr = repo.RenameApiResponse
	return
}
