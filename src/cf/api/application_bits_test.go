package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	testapi "testhelpers/api"
	testcf "testhelpers/cf"
	testnet "testhelpers/net"
	"testing"
)

var expectedResources = testapi.RemoveWhiteSpaceFromBody(`[
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

var uploadApplicationRequest = testnet.TestRequest{
	Method: "PUT",
	Path: "/v2/apps/my-cool-app-guid/bits",
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
}


var matchResourceRequest = testnet.TestRequest {
	Method:    "PUT",
	Path: "/v2/resource_match",
	Matcher: testnet.RequestBodyMatcher(expectedResources),
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body: `[
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
]`,
	},
}

var defaultRequests = []testnet.TestRequest{
	matchResourceRequest,
	uploadApplicationRequest,
	createProgressEndpoint("running"),
	createProgressEndpoint("finished"),
}

var uploadBodyMatcher = func(request *http.Request) bool {
	bodyBytes, err := ioutil.ReadAll(request.Body)

	if err != nil {
		return false
	}

	bodyString := string(bodyBytes)
	zipAttachmentContentDispositionMatches := strings.Contains(bodyString, `Content-Disposition: form-data; name="application"; filename="application.zip"`)
	if !zipAttachmentContentDispositionMatches {
		println("Zip Attachment ContentDisposition does not match")
	}
	zipAttachmentContentTypeMatches := strings.Contains(bodyString, `Content-Type: application/zip`)
	if !zipAttachmentContentTypeMatches {
		println("Zip Attachment Content Type does not match")
	}
	zipAttachmentContentTransferEncodingMatches := strings.Contains(bodyString, `Content-Transfer-Encoding: binary`)
	if !zipAttachmentContentTransferEncodingMatches {
		println("Zip Attachment Content Transfer Encoding does not match")
	}
	zipAttachmentContentLengthPresent := strings.Contains(bodyString, `Content-Length:`)
	if !zipAttachmentContentLengthPresent {
		println("Zip Attachment Content Length missing")
	}
	zipAttachmentContentPresent := strings.Contains(bodyString, `hello world!`)
	if !zipAttachmentContentPresent {
		println("Zip Attachment Content missing")
	}

	resourcesContentDispositionMatches := strings.Contains(bodyString, `Content-Disposition: form-data; name="resources"`)
	if !resourcesContentDispositionMatches {
		println("Resources Content Disposition does not match")
	}
	resourcesPresent := strings.Contains(bodyString, expectedResources)
	if !resourcesPresent {
		println("Resources not present")
	}

	return zipAttachmentContentDispositionMatches &&
			zipAttachmentContentTypeMatches &&
			zipAttachmentContentTransferEncodingMatches &&
			zipAttachmentContentLengthPresent &&
			zipAttachmentContentPresent &&
			resourcesContentDispositionMatches &&
			resourcesPresent
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
		Body:body,
	}

	return
}

func TestUploadWithInvalidDirectory(t *testing.T) {
	config := &configuration.Configuration{}
	gateway := net.NewCloudControllerGateway()
	zipper := &testcf.FakeZipper{}

	repo := NewCloudControllerApplicationBitsRepository(config, gateway, zipper)
	app := cf.Application{}

	apiResponse := repo.UploadApp(app, "/foo/bar")
	assert.True(t, apiResponse.IsNotSuccessful())
	assert.Contains(t, apiResponse.Message, "Error listing app files")
}

func TestUploadApp(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	dir = filepath.Join(dir, "../../fixtures/example-app")

	app, apiResponse := testUploadApp(t, dir, defaultRequests)
	assert.True(t, apiResponse.IsSuccessful())
	testUploadDir(t,app)
}

func TestCreateUploadDirWithAZipFile(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	dir = filepath.Join(dir, "../../fixtures/example-app.zip")

	app, apiResponse := testUploadApp(t, dir, defaultRequests)
	assert.True(t, apiResponse.IsSuccessful())
	testUploadDir(t,app)
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

func testUploadApp(t *testing.T, dir string, requests []testnet.TestRequest) (app cf.Application, apiResponse net.ApiResponse){
	ts, handler := testnet.NewServer(t, requests)
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway()
	zipper := &testcf.FakeZipper{ZippedBuffer: bytes.NewBufferString("hello world!")}
	repo := NewCloudControllerApplicationBitsRepository(config, gateway, zipper)

	app = cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	apiResponse = repo.UploadApp(app, dir)

	assert.True(t,handler.AllRequestsCalled())

	return
}

func testUploadDir(t *testing.T, app cf.Application){
	uploadDir := cf.TempDirForApp(app)
	files, err := filepath.Glob(filepath.Join(uploadDir, "*"))
	assert.NoError(t, err)

	assert.Equal(t, files, []string{
			filepath.Join(uploadDir, "Gemfile"),
			filepath.Join(uploadDir, "Gemfile.lock"),
			filepath.Join(uploadDir, "manifest.yml"),
		})
}
