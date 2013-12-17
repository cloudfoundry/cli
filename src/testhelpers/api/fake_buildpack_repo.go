package api

import (
	"cf"
	"cf/net"
)

type FakeBuildpackRepository struct {
	Buildpacks []cf.Buildpack

	FindByNameNotFound    bool
	FindByNameName        string
	FindByNameBuildpack   cf.Buildpack
	FindByNameApiResponse net.ApiResponse

	CreateBuildpackExists bool
	CreateBuildpack       cf.Buildpack
	CreateApiResponse     net.ApiResponse

	DeleteBuildpackGuid string
	DeleteApiResponse   net.ApiResponse

	UpdateBuildpack cf.Buildpack
}

func (repo *FakeBuildpackRepository) ListBuildpacks(stop chan bool) (buildpacksChan chan []cf.Buildpack, statusChan chan net.ApiResponse) {
	buildpacksChan = make(chan []cf.Buildpack, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		buildpackCount := len(repo.Buildpacks)
		for i := 0; i < buildpackCount; i += 2 {
			select {
			case <-stop:
				break
			default:
				if buildpackCount-i > 1 {
					buildpacksChan <- repo.Buildpacks[i : i+2]
				} else {
					buildpacksChan <- repo.Buildpacks[i:]
				}
			}
		}

		close(buildpacksChan)
		close(statusChan)

		cf.WaitForClose(stop)
	}()

	return
}

func (repo *FakeBuildpackRepository) FindByName(name string) (buildpack cf.Buildpack, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	buildpack = repo.FindByNameBuildpack

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("Buildpack %s not found", name)
	}

	return
}

func (repo *FakeBuildpackRepository) Create(name string, position *int, enabled *bool) (createdBuildpack cf.Buildpack, apiResponse net.ApiResponse) {
	if repo.CreateBuildpackExists {
		return repo.CreateBuildpack, net.NewApiResponse("Buildpack already exists", cf.BUILDPACK_EXISTS, 400)
	}

	repo.CreateBuildpack = cf.Buildpack{BasicFields: cf.BasicFields{Name: name}, Position: position, Enabled: enabled}
	return repo.CreateBuildpack, repo.CreateApiResponse
}

func (repo *FakeBuildpackRepository) Delete(buildpackGuid string) (apiResponse net.ApiResponse) {
	repo.DeleteBuildpackGuid = buildpackGuid
	apiResponse = repo.DeleteApiResponse
	return
}

func (repo *FakeBuildpackRepository) Update(buildpack cf.Buildpack) (updatedBuildpack cf.Buildpack, apiResponse net.ApiResponse) {
	repo.UpdateBuildpack = buildpack
	return
}
