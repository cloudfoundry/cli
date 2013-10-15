package api

import (
	"archive/zip"
	"cf"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	testnet "testhelpers/net"
	"testing"
)

var uploadBuildpackRequest = testnet.TestRequest{
	Method:  "PUT",
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
var uploadBuildpackBodyMatcher = func(request *http.Request) bool {
	err := request.ParseMultipartForm(4096)
	defer request.MultipartForm.RemoveAll()
	if err != nil {
		println("Invalid request")
		return false
	}

	if len(request.MultipartForm.Value) != 0 {
		println("Should have 0 values")
		return false
	}

	if len(request.MultipartForm.File) != 1 {
		println("Wrong number of files")
		return false
	}

	for k, v := range request.MultipartForm.File {
		if k != "buildpack" && len(v) == 1 && v[0].Filename != "buildpack.zip" {
			println("Wrong content disposition")
			return false
		}
		multipartFile := v[0]

		var file multipart.File
		if file, err = multipartFile.Open(); err != nil {
			println("Cannot get multipart file")
			return false
		}

		if zipReader, err := zip.NewReader(file, 4096); err != nil {
			println("Error reading zip content")
			return false
		} else {
			if len(zipReader.File) != 3 {
				println("Wrong number of files in zip")
				return false
			}

		nextFile:
			for _, f := range zipReader.File {
				for _, expected := range buildpackContent {
					if f.Name == expected {
						continue nextFile
					}
				}
				return false
			}

		}
	}

	return true
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
	ts, handler := testnet.NewServer(t, requests)
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
