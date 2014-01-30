package api_test

import (
	"archive/zip"
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"fileutils"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
	"testing"
	"time"
)

var permissionsToSet os.FileMode
var expectedPermissionBits os.FileMode

func init() {
	permissionsToSet = 0467
	fileutils.TempFile("permissionedFile", func(file *os.File, err error) {
		if err != nil {
			panic("could not create tmp file")
		}

		fileInfo, err := file.Stat()
		if err != nil {
			panic("could not stat tmp file")
		}

		expectedPermissionBits = fileInfo.Mode()
		if runtime.GOOS != "windows" {
			expectedPermissionBits |= permissionsToSet & 0111
		}
	})
}

var uploadApplicationRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method:  "PUT",
	Path:    "/v2/apps/my-cool-app-guid/bits",
	Matcher: uploadBodyMatcher,
	Response: testnet.TestResponse{
		Status: http.StatusCreated,
		Body: `
{
	"metadata":{
		"guid": "my-job-guid",
		"url": "/v2/jobs/my-job-guid"
	}
}
	`},
})
var defaultRequests = []testnet.TestRequest{
	uploadApplicationRequest,
	createProgressEndpoint("running"),
	createProgressEndpoint("finished"),
}

var expectedApplicationContent = []string{"Gemfile", "Gemfile.lock", "manifest.yml", "app.rb", "config.ru"}

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
	assert.Equal(t, chompedResourceManifest, "[]", "Resources do not match")

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

	assert.Equal(t, len(zipReader.File), 5, "Wrong number of files in zip")
	assert.Equal(t, zipReader.File[0].Mode(), uint32(expectedPermissionBits))

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

	apiResponse := repo.UploadApp("app-guid", "/foo/bar", func(path string, uploadSize, fileCount uint64) {})
	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Contains(t, apiResponse.Message, filepath.Join("foo", "bar"))
}

func TestUploadApp(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	dir = filepath.Join(dir, "../../fixtures/example-app")
	err = os.Chmod(filepath.Join(dir, "Gemfile"), permissionsToSet)

	assert.NoError(t, err)

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

func TestCreateUploadDirWithAZipLikeFile(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	dir = filepath.Join(dir, "../../fixtures/example-app.azip")

	_, apiResponse := testUploadApp(t, dir, defaultRequests)
	assert.True(t, apiResponse.IsSuccessful())
}

func TestUploadAppFailsWhilePushingBits(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	dir = filepath.Join(dir, "../../fixtures/example-app")

	requests := []testnet.TestRequest{
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
	gateway.PollingThrottle = time.Duration(0)
	zipper := cf.ApplicationZipper{}
	repo := NewCloudControllerApplicationBitsRepository(config, gateway, zipper)

	var (
		reportedPath                          string
		reportedFileCount, reportedUploadSize uint64
	)
	apiResponse = repo.UploadApp("my-cool-app-guid", dir, func(path string, uploadSize, fileCount uint64) {
		reportedPath = path
		reportedUploadSize = uploadSize
		reportedFileCount = fileCount
	})

	assert.Equal(t, reportedPath, dir)
	assert.Equal(t, reportedFileCount, uint64(len(expectedApplicationContent)))
	assert.Equal(t, reportedUploadSize, uint64(1094))
	assert.True(t, handler.AllRequestsCalled())

	return
}
