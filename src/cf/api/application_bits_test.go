package api_test

import (
	"bytes"
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testhelpers"
	"testing"
)

var expectedResources = testhelpers.RemoveWhiteSpaceFromBody(`[
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

var uploadApplicationEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/apps/my-cool-app-guid/bits",
	uploadBodyMatcher,
	testhelpers.TestResponse{Status: http.StatusCreated},
)

var matchResourcesEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/resource_match",
	testhelpers.RequestBodyMatcher(expectedResources),
	testhelpers.TestResponse{
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
]`},
)

var uploadEndpoints = func(writer http.ResponseWriter, request *http.Request) {
	if strings.Contains(request.URL.Path, "bits") {
		uploadApplicationEndpoint(writer, request)
		return
	}

	matchResourcesEndpoint(writer, request)
}

func TestUploadApp(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	dir = filepath.Join(dir, "../../fixtures/example-app")

	testUploadApp(t, dir)
}

func TestCreateUploadDirWithAZipFile(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	dir = filepath.Join(dir, "../../fixtures/example-app.zip")

	testUploadApp(t, dir)
}

func testUploadApp(t *testing.T, dir string) {
	ts := httptest.NewTLSServer(http.HandlerFunc(uploadEndpoints))
	defer ts.Close()

	config := configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	zipper := &testhelpers.FakeZipper{ZippedBuffer: bytes.NewBufferString("hello world!")}
	repo := NewCloudControllerApplicationBitsRepository(config, gateway, zipper)

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	apiErr := repo.UploadApp(app, dir)
	assert.NoError(t, apiErr)

	uploadDir := cf.TempDirForApp(app)
	files, err := filepath.Glob(filepath.Join(uploadDir, "*"))
	assert.NoError(t, err)

	assert.Equal(t, files, []string{
		filepath.Join(uploadDir, "Gemfile"),
		filepath.Join(uploadDir, "Gemfile.lock"),
		filepath.Join(uploadDir, "manifest.yml"),
	})
}
