package api_test

import (
	"archive/zip"
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"os"
	"path/filepath"
	testnet "testhelpers/net"
)

func uploadBuildpackRequest(filename string) testnet.TestRequest {
	return testnet.TestRequest{
		Method:  "PUT",
		Path:    "/v2/buildpacks/my-cool-buildpack-guid/bits",
		Matcher: uploadBuildpackBodyMatcher(filename),
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
}

var expectedBuildpackContent = []string{"detect", "compile", "package"}

func uploadBuildpackBodyMatcher(pathToFile string) testnet.RequestMatcher {
	return func(t mr.TestingT, request *http.Request) {
		err := request.ParseMultipartForm(4096)
		if err != nil {
			assert.Fail(t, "Failed parsing multipart form: %s", err)
			return
		}
		defer request.MultipartForm.RemoveAll()

		assert.Equal(t, len(request.MultipartForm.Value), 0, "Should have 0 values")
		assert.Equal(t, len(request.MultipartForm.File), 1, "Wrong number of files")

		files, ok := request.MultipartForm.File["buildpack"]

		assert.True(t, ok, "Buildpack file part not present")
		assert.Equal(t, len(files), 1, "Wrong number of files")

		buildpackFile := files[0]
		assert.Equal(t, buildpackFile.Filename, filepath.Base(pathToFile), "Wrong file name")

		file, err := buildpackFile.Open()
		if err != nil {
			assert.Fail(t, "Cannot get multipart file: %s", err.Error())
			return
		}

		zipReader, err := zip.NewReader(file, 4096)
		if err != nil {
			assert.Fail(t, "Error reading zip content: %s", err.Error())
		}

		assert.Equal(t, len(zipReader.File), 3, "Wrong number of files in zip")
		assert.Equal(t, zipReader.File[1].Mode(), uint32(0666))

	nextFile:
		for _, f := range zipReader.File {
			for _, expected := range expectedBuildpackContent {
				if f.Name == expected {
					continue nextFile
				}
			}
			assert.Fail(t, "Missing file: "+f.Name)
		}
	}
}

func testUploadBuildpack(t mr.TestingT, dir string, requests []testnet.TestRequest) (buildpack models.Buildpack, apiResponse net.ApiResponse) {
	ts, handler := testnet.NewTLSServer(t, requests)
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerBuildpackBitsRepository(config, gateway, cf.ApplicationZipper{})
	buildpack = models.Buildpack{Name: "my-cool-buildpack", Guid: "my-cool-buildpack-guid"}

	apiResponse = repo.UploadBuildpack(buildpack, dir)
	assert.True(t, handler.AllRequestsCalled())
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestUploadBuildpackWithInvalidDirectory", func() {
			config := &configuration.Configuration{}
			gateway := net.NewCloudControllerGateway()

			repo := NewCloudControllerBuildpackBitsRepository(config, gateway, cf.ApplicationZipper{})
			buildpack := models.Buildpack{}

			apiResponse := repo.UploadBuildpack(buildpack, "/foo/bar")
			assert.True(mr.T(), apiResponse.IsNotSuccessful())
			assert.Contains(mr.T(), apiResponse.Message, "Invalid buildpack")
		})
		It("TestUploadBuildpack", func() {

			dir, err := os.Getwd()
			assert.NoError(mr.T(), err)
			dir = filepath.Join(dir, "../../fixtures/example-buildpack")
			err = os.Chmod(filepath.Join(dir, "detect"), 0666)
			assert.NoError(mr.T(), err)

			_, apiResponse := testUploadBuildpack(mr.T(), dir, []testnet.TestRequest{
				uploadBuildpackRequest(dir),
			})
			assert.True(mr.T(), apiResponse.IsSuccessful())
		})
		It("TestUploadBuildpackWithAZipFile", func() {

			dir, err := os.Getwd()
			assert.NoError(mr.T(), err)
			dir = filepath.Join(dir, "../../fixtures/example-buildpack.zip")

			_, apiResponse := testUploadBuildpack(mr.T(), dir, []testnet.TestRequest{
				uploadBuildpackRequest(dir),
			})
			assert.True(mr.T(), apiResponse.IsSuccessful())
		})
	})
}
