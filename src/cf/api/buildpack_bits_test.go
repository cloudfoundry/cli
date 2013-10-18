package api

import (
	"archive/zip"
	"cf"
	"cf/configuration"
	"cf/net"
	"errors"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	testnet "testhelpers/net"
	"testing"
)

var uploadBuildpackRequest = testnet.TestRequest{
	Method:  "POST",
	Path:    "/v2/buildpacks/my-cool-buildpack-guid/bits",
	Matcher: uploadBuildpackBodyMatcher,
	Response: testnet.TestResponse{
		Status: http.StatusCreated,
		Body: `
{
	"metadata":{
		"guid": "my-job-guid"
	}
}
	`},
}

var buildpackContent = []string{"detect", "compile", "package"}
var uploadBuildpackBodyMatcher = func(request *http.Request) error {
	err := request.ParseMultipartForm(4096)
	defer request.MultipartForm.RemoveAll()
	if err != nil {
		return err
	}

	if len(request.MultipartForm.Value) != 0 {
		return errors.New("Should have 0 values")
	}

	if len(request.MultipartForm.File) != 1 {
		return errors.New("Wrong number of files")
	}

	for k, v := range request.MultipartForm.File {
		if k != "buildpack" && len(v) == 1 && v[0].Filename != "buildpack.zip" {
			return errors.New("Wrong content disposition")
		}
		multipartFile := v[0]

		var file multipart.File
		if file, err = multipartFile.Open(); err != nil {
			return errors.New("Cannot get multipart file")

		}

		if zipReader, err := zip.NewReader(file, 4096); err != nil {
			return errors.New("Error reading zip content")
		} else {
			if len(zipReader.File) != 3 {
				return errors.New("Wrong number of files in zip")
			}

		nextFile:
			for _, f := range zipReader.File {
				for _, expected := range buildpackContent {
					if f.Name == expected {
						continue nextFile
					}
				}
				return errors.New("Missing file: " + f.Name)
			}

		}
	}

	return nil
}

var defaultBuildpackRequests = []testnet.TestRequest{
	uploadBuildpackRequest,
}

func TestUploadBuildpackWithInvalidDirectory(t *testing.T) {
	config := &configuration.Configuration{}
	gateway := net.NewCloudControllerGateway()

	repo := NewCloudControllerBuildpackBitsRepository(config, gateway, cf.ApplicationZipper{})
	buildpack := cf.Buildpack{}

	apiResponse := repo.UploadBuildpack(buildpack, "/foo/bar")
	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Contains(t, apiResponse.Message, "Invalid buildpack")
}

func TestUploadBuildpack(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	dir = filepath.Join(dir, "../../fixtures/example-buildpack")

	_, apiResponse := testUploadBuildpack(t, dir, defaultBuildpackRequests)
	assert.True(t, apiResponse.IsSuccessful())
}

func TestUploadBuildpackWithAZipFile(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	dir = filepath.Join(dir, "../../fixtures/example-buildpack.zip")

	_, apiResponse := testUploadBuildpack(t, dir, defaultBuildpackRequests)
	assert.True(t, apiResponse.IsSuccessful())
}

func testUploadBuildpack(t *testing.T, dir string, requests []testnet.TestRequest) (buildpack cf.Buildpack, apiResponse net.ApiResponse) {
	ts, handler := testnet.NewTLSServer(t, requests)
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerBuildpackBitsRepository(config, gateway, cf.ApplicationZipper{})

	buildpack = cf.Buildpack{Name: "my-cool-buildpack", Guid: "my-cool-buildpack-guid"}

	apiResponse = repo.UploadBuildpack(buildpack, dir)
	assert.True(t, handler.AllRequestsCalled())
	return
}
