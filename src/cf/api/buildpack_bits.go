package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

type BuildpackBitsRepository interface {
	UploadBuildpack(buildpack cf.Buildpack, dir string) (apiResponse net.ApiResponse)
}

type CloudControllerBuildpackBitsRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
	zipper  cf.Zipper
}

func NewCloudControllerBuildpackBitsRepository(config *configuration.Configuration, gateway net.Gateway, zipper cf.Zipper) (repo CloudControllerBuildpackBitsRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.zipper = zipper
	return
}

func (repo CloudControllerBuildpackBitsRepository) UploadBuildpack(buildpack cf.Buildpack, dir string) (apiResponse net.ApiResponse) {
	zipFile, err := repo.zipper.Zip(dir)
	if err != nil {
		return net.NewApiResponseWithError("Invalid buildpack", err)
	}
	defer zipFile.Close()

	apiResponse = repo.uploadBits(buildpack, zipFile)
	if apiResponse.IsNotSuccessful() {
		return
	}

	return
}

func (repo CloudControllerBuildpackBitsRepository) uploadBits(app cf.Buildpack, zipFile *os.File) (apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/buildpacks/%s/bits", repo.config.Target, app.Guid)

	body, boundary, err := createBuildpackUploadBody(zipFile)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Error creating upload", err)
		return
	}

	request, apiResponse := repo.gateway.NewRequest("PUT", url, repo.config.AccessToken, body)
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
	request.HttpReq.Header.Set("Content-Type", contentType)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func createBuildpackUploadBody(zipFile *os.File) (body *os.File, boundary string, err error) {
	body, err = os.Create(cf.TempFileForRequestBody())
	if err != nil {
		return
	}

	writer := multipart.NewWriter(body)
	defer writer.Close()

	boundary = writer.Boundary()

	zipStats, err := zipFile.Stat()
	if err != nil {
		return
	}

	if zipStats.Size() == 0 {
		return
	}

	part, err := writer.CreateFormFile("buildpack", "buildpack.zip")
	if err != nil {
		return
	}

	_, err = io.Copy(part, zipFile)
	if err != nil {
		return
	}

	return
}
