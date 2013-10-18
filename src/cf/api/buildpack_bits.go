package api

import (
    "bytes"
    "cf"
    "cf/configuration"
    "cf/net"
    "fmt"
    "io"
    "mime/multipart"
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
    zipBuffer, err := repo.zipper.Zip(dir)
    if err != nil {
        return net.NewApiResponseWithError("Invalid buildpack", err)
    }

    apiResponse = repo.uploadBits(buildpack, zipBuffer)
    if apiResponse.IsNotSuccessful() {
        return
    }

    return
}

func (repo CloudControllerBuildpackBitsRepository) uploadBits(app cf.Buildpack, zipBuffer *bytes.Buffer) (apiResponse net.ApiResponse) {
    url := fmt.Sprintf("%s/v2/buildpacks/%s/bits", repo.config.Target, app.Guid)

    body, boundary, err := createBuildpackUploadBody(zipBuffer)
    if err != nil {
        apiResponse = net.NewApiResponseWithError("Error creating upload", err)
        return
    }

    request, apiResponse := repo.gateway.NewRequest("POST", url, repo.config.AccessToken, body)
    contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
    request.Header.Set("Content-Type", contentType)
    if apiResponse.IsNotSuccessful() {
        return
    }

    apiResponse = repo.gateway.PerformRequest(request)
    return
}

func createBuildpackUploadBody(zipBuffer *bytes.Buffer) (body *bytes.Buffer, boundary string, err error) {
    body = new(bytes.Buffer)

    writer := multipart.NewWriter(body)
    defer writer.Close()

    boundary = writer.Boundary()

    if zipBuffer.Len() == 0 {
        return
    }

    part, err := writer.CreateFormFile("buildpack", "buildpack.zip")
    if err != nil {
        return
    }

    _, err = io.Copy(part, zipBuffer)
    if err != nil {
        return
    }

    return
}
