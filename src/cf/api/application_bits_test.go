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
	"strings"
	"testhelpers"
	"testing"
)

var uploadBodyMatcher = func(request *http.Request) bool {
	bodyBytes, err := ioutil.ReadAll(request.Body)

	if err != nil {
		return false
	}

	bodyString := string(bodyBytes)
	zipAttachmentContentDispositionMatches := strings.Contains(bodyString, `Content-Disposition: form-data; name="application"; filename="application.zip"`)
	zipAttachmentContentTypeMatches := strings.Contains(bodyString, `Content-Type: application/zip`)
	zipAttachmentContentTransferEncodingMatches := strings.Contains(bodyString, `Content-Transfer-Encoding: binary`)
	zipAttachmentContentLengthPresent := strings.Contains(bodyString, `Content-Length:`)
	zipAttachmentContentPresent := strings.Contains(bodyString, `hello world!`)

	resourcesContentDispositionMatches := strings.Contains(bodyString, `Content-Disposition: form-data; name="resources"`)

	return zipAttachmentContentDispositionMatches &&
		zipAttachmentContentTypeMatches &&
		zipAttachmentContentTransferEncodingMatches &&
		zipAttachmentContentLengthPresent &&
		zipAttachmentContentPresent &&
		resourcesContentDispositionMatches
}

var uploadApplicationEndpoint = testhelpers.CreateEndpoint(
	"PUT",
	"/v2/apps/my-cool-app-guid/bits",
	uploadBodyMatcher,
	testhelpers.TestResponse{Status: http.StatusCreated},
)

func TestUploadApplication(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(uploadApplicationEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}
	gateway := net.NewCloudControllerGateway(&testhelpers.FakeAuthenticator{})
	repo := NewCloudControllerApplicationBitsRepository(config, gateway)

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}
	zipBuffer := bytes.NewBufferString("hello world!")

	err := repo.Upload(app, zipBuffer)
	assert.NoError(t, err)
}
