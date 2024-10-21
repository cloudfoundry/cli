package apifakes

import (
	"code.cloudfoundry.org/cli/v9/cf/errors"
	"code.cloudfoundry.org/cli/v9/cf/models"
)

type OldFakeBuildpackBitsRepository struct {
	UploadBuildpackErr         bool
	UploadBuildpackAPIResponse error
	UploadBuildpackPath        string
}

func (repo *OldFakeBuildpackBitsRepository) UploadBuildpack(buildpack models.Buildpack, dir string) error {
	if repo.UploadBuildpackErr {
		return errors.New("Invalid buildpack")
	}

	repo.UploadBuildpackPath = dir
	return repo.UploadBuildpackAPIResponse
}
