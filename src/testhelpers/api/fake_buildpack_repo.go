package api

import (
	"cf"
	"cf/models"
	"cf/net"
)

type FakeBuildpackRepository struct {
	Buildpacks []models.Buildpack

	FindByNameNotFound    bool
	FindByNameName        string
	FindByNameBuildpack   models.Buildpack
	FindByNameApiResponse net.ApiResponse

	CreateBuildpackExists bool
	CreateBuildpack       models.Buildpack
	CreateApiResponse     net.ApiResponse

	DeleteBuildpackGuid string
	DeleteApiResponse   net.ApiResponse

	UpdateBuildpack models.Buildpack
}

func (repo *FakeBuildpackRepository) ListBuildpacks(cb func([]models.Buildpack) bool) net.ApiResponse {
	if len(repo.Buildpacks) > 0 {
		cb(repo.Buildpacks)
	}
	return net.NewApiResponseWithStatusCode(200)
}

func (repo *FakeBuildpackRepository) FindByName(name string) (buildpack models.Buildpack, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	buildpack = repo.FindByNameBuildpack

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("Buildpack %s not found", name)
	}

	return
}

func (repo *FakeBuildpackRepository) Create(name string, position *int, enabled *bool, locked *bool) (createdBuildpack models.Buildpack, apiResponse net.ApiResponse) {
	if repo.CreateBuildpackExists {
		return repo.CreateBuildpack, net.NewApiResponse("Buildpack already exists", cf.BUILDPACK_EXISTS, 400)
	}

	repo.CreateBuildpack = models.Buildpack{Name: name, Position: position, Enabled: enabled, Locked: locked}
	return repo.CreateBuildpack, repo.CreateApiResponse
}

func (repo *FakeBuildpackRepository) Delete(buildpackGuid string) (apiResponse net.ApiResponse) {
	repo.DeleteBuildpackGuid = buildpackGuid
	apiResponse = repo.DeleteApiResponse
	return
}

func (repo *FakeBuildpackRepository) Update(buildpack models.Buildpack) (updatedBuildpack models.Buildpack, apiResponse net.ApiResponse) {
	repo.UpdateBuildpack = buildpack
	return
}
