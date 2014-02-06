package api

import (
"cf/models"
	"cf/net"
)

type FakeBuildpackBitsRepository struct {
	UploadBuildpackErr         bool
	UploadBuildpackApiResponse net.ApiResponse
	UploadBuildpackPath        string
}

func (repo *FakeBuildpackBitsRepository) UploadBuildpack(buildpack models.Buildpack, dir string) net.ApiResponse {
	if repo.UploadBuildpackErr {
		return net.NewApiResponseWithMessage("Invalid buildpack")
	}

	repo.UploadBuildpackPath = dir
	return repo.UploadBuildpackApiResponse
}
