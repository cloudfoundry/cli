package api

import (
	"archive/zip"
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
	"testing"
)

var expectedResources = testnet.RemoveWhiteSpaceFromBody(`[
    {
        "fn": "Gemfile",
        "sha1": "d9c3a51de5c89c11331d3b90b972789f1a14699a",
        "size": 59
    },
    {
        "fn": "Gemfile.lock",
        "sha1": "345f999aef9070fb9a608e65cf221b7038156b6d",
        "size": 229
    },
    {
        "fn": "app.rb",
        "sha1": "2474735f5163ba7612ef641f438f4b5bee00127b",
        "size": 51
    },
    {
        "fn": "config.ru",
        "sha1": "f097424ce1fa66c6cb9f5e8a18c317376ec12e05",
        "size": 70
    },
    {
        "fn": "manifest.yml",
        "sha1": "19b5b4225dc64da3213b1ffaa1e1920ee5faf36c",
        "size": 111
    }
]`)

var matchedResources = testnet.RemoveWhiteSpaceFromBody(`[
	{
        "fn": "app.rb",
        "sha1": "2474735f5163ba7612ef641f438f4b5bee00127b",
        "size": 51
    },
    {
        "fn": "config.ru",
        "sha1": "f097424ce1fa66c6cb9f5e8a18c317376ec12e05",
        "size": 70
    }
]`)

var uploadApplicationRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method:  "PUT",
	Path:    "/v2/apps/my-cool-app-guid/bits",
	Matcher: uploadBodyMatcher,
	Response: testnet.TestResponse{
		Status: http.StatusCreated,
		Body: `
{
	"metadata":{
		"guid": "my-job-guid"
	}
}
	`},
})

var matchResourceRequest = testnet.TestRequest{
	Method:  "PUT",
	Path:    "/v2/resource_match",
	Matcher: testnet.RequestBodyMatcher(expectedResources),
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body:   matchedResources,
	},
}

var defaultRequests = []testnet.TestRequest{
	matchResourceRequest,
	uploadApplicationRequest,
	createProgressEndpoint("running"),
	createProgressEndpoint("finished"),
}

var expectedApplicationContent = []string{"Gemfile", "Gemfile.lock", "manifest.yml"}

var uploadBodyMatcher = func(t *testing.T, request *http.Request) {
	err := request.ParseMultipartForm(4096)
	if err != nil {
		assert.Fail(t, "Failed parsing multipart form", err)
		return
	}
	defer request.MultipartForm.RemoveAll()

	assert.Equal(t, len(request.MultipartForm.Value), 1, "Should have 1 value")
	valuePart, ok := request.MultipartForm.Value["resources"]
	assert.True(t, ok, "Resource manifest not present")
	assert.Equal(t, len(valuePart), 1, "Wrong number of values")

	resourceManifest := valuePart[0]
	chompedResourceManifest := strings.Replace(resourceManifest, "\n", "", -1)
	assert.Equal(t, chompedResourceManifest, matchedResources, "Resources do not match")

	assert.Equal(t, len(request.MultipartForm.File), 1, "Wrong number of files")

	fileHeaders, ok := request.MultipartForm.File["application"]
	assert.True(t, ok, "Application file part not present")
	assert.Equal(t, len(fileHeaders), 1, "Wrong number of files")

	applicationFile := fileHeaders[0]
	assert.Equal(t, applicationFile.Filename, "application.zip", "Wrong file name")

	file, err := applicationFile.Open()
	if err != nil {
		assert.Fail(t, "Cannot get multipart file", err.Error())
		return
	}

	length, err := strconv.ParseInt(applicationFile.Header.Get("content-length"), 10, 64)
	if err != nil {
		assert.Fail(t, "Cannot convert content-length to int", err.Error())
		return
	}

	zipReader, err := zip.NewReader(file, length)
	if err != nil {
		assert.Fail(t, "Error reading zip content", err.Error())
		return
	}

	assert.Equal(t, len(zipReader.File), 3, "Wrong number of files in zip")

nextFile:
	for _, f := range zipReader.File {
		for _, expected := range expectedApplicationContent {
			if f.Name == expected {
				continue nextFile
			}
		}
		assert.Fail(t, "Missing file: "+f.Name)
	}
}

func createProgressEndpoint(status string) (req testnet.TestRequest) {
	body := fmt.Sprintf(`
	{
		"entity":{
			"status":"%s"
		}
	}`, status)

	req.Method = "GET"
	req.Path = "/v2/jobs/my-job-guid"
	req.Response = testnet.TestResponse{
		Status: http.StatusCreated,
		Body:   body,
	}

	return
}

func TestUploadWithInvalidDirectory(t *testing.T) {
	config := &configuration.Configuration{}
	gateway := net.NewCloudControllerGateway()
	zipper := &cf.ApplicationZipper{}

	repo := NewCloudControllerApplicationBitsRepository(config, gateway, zipper)

	apiResponse := repo.UploadApp("app-guid", "/foo/bar")
	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Contains(t, apiResponse.Message, "/foo/bar")
}

func TestUploadApp(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	dir = filepath.Join(dir, "../../fixtures/example-app")

	_, apiResponse := testUploadApp(t, dir, defaultRequests)
	assert.True(t, apiResponse.IsSuccessful())
}

func TestCreateUploadDirWithAZipFile(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	dir = filepath.Join(dir, "../../fixtures/example-app.zip")

	_, apiResponse := testUploadApp(t, dir, defaultRequests)
	assert.True(t, apiResponse.IsSuccessful())
}

func TestUploadAppFailsWhilePushingBits(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	dir = filepath.Join(dir, "../../fixtures/example-app")

	requests := []testnet.TestRequest{
		matchResourceRequest,
		uploadApplicationRequest,
		createProgressEndpoint("running"),
		createProgressEndpoint("failed"),
	}
	_, apiResponse := testUploadApp(t, dir, requests)
	assert.False(t, apiResponse.IsSuccessful())
}

func testUploadApp(t *testing.T, dir string, requests []testnet.TestRequest) (app cf.Application, apiResponse net.ApiResponse) {
	ts, handler := testnet.NewTLSServer(t, requests)
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	zipper := cf.ApplicationZipper{}
	repo := NewCloudControllerApplicationBitsRepository(config, gateway, zipper)

	apiResponse = repo.UploadApp("my-cool-app-guid", dir)
	assert.True(t, handler.AllRequestsCalled())

	return
}
