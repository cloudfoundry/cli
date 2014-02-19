package api

import (
	"cf"
	"cf/configuration"
	"cf/models"
	"cf/net"
	"crypto/tls"
	"fileutils"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type BuildpackBitsRepository interface {
	UploadBuildpack(buildpack models.Buildpack, dir string) (apiResponse net.ApiResponse)
}

type CloudControllerBuildpackBitsRepository struct {
	config  configuration.Reader
	gateway net.Gateway
	zipper  cf.Zipper
}

func NewCloudControllerBuildpackBitsRepository(config configuration.Reader, gateway net.Gateway, zipper cf.Zipper) (repo CloudControllerBuildpackBitsRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.zipper = zipper
	return
}

func (repo CloudControllerBuildpackBitsRepository) UploadBuildpack(buildpack models.Buildpack, dir string) (apiResponse net.ApiResponse) {
	if strings.HasPrefix(dir, "http://") || strings.HasPrefix(dir, "https://") {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				Proxy:           http.ProxyFromEnvironment,
			},
		}
		response, err := client.Get(dir)
		if err != nil {
			apiResponse = net.NewApiResponseWithError("Could not download buildpack", err)
		} else {
			apiResponse = repo.uploadBits(buildpack, response.Body, path.Base(dir))
		}
	} else {
		fileutils.TempFile("buildpack", func(zipFile *os.File, err error) {
			if err != nil {
				apiResponse = net.NewApiResponseWithMessage(err.Error())
				return
			}

			err = repo.zipper.Zip(dir, zipFile)
			if err != nil {
				apiResponse = net.NewApiResponseWithError("Invalid buildpack", err)
				return
			}
			apiResponse = repo.uploadBits(buildpack, zipFile, filepath.Base(dir))
		})
	}

	return
}

func (repo CloudControllerBuildpackBitsRepository) uploadBits(buildpack models.Buildpack, body io.Reader, buildpackName string) net.ApiResponse {
	return repo.performMultiPartUpload(
		fmt.Sprintf("%s/v2/buildpacks/%s/bits", repo.config.ApiEndpoint(), buildpack.Guid),
		"buildpack",
		buildpackName,
		body)
}

func (repo CloudControllerBuildpackBitsRepository) performMultiPartUpload(url string, fieldName string, fileName string, body io.Reader) (apiResponse net.ApiResponse) {
	fileutils.TempFile("requests", func(requestFile *os.File, err error) {
		if err != nil {
			apiResponse = net.NewApiResponseWithMessage(err.Error())
			return
		}

		writer := multipart.NewWriter(requestFile)
		part, err := writer.CreateFormFile(fieldName, fileName)

		if err != nil {
			writer.Close()
			return
		}

		_, err = io.Copy(part, body)
		writer.Close()

		if err != nil {
			apiResponse = net.NewApiResponseWithError("Error creating upload", err)
			return
		}

		var request *net.Request
		request, apiResponse = repo.gateway.NewRequest("PUT", url, repo.config.AccessToken(), requestFile)
		contentType := fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary())
		request.HttpReq.Header.Set("Content-Type", contentType)
		if apiResponse.IsNotSuccessful() {
			return
		}

		apiResponse = repo.gateway.PerformRequest(request)
	})

	return
}
