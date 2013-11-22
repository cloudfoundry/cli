package api

import (
	"cf/configuration"
	"cf/net"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testnet "testhelpers/net"
	"testing"
)

func TestListFiles(t *testing.T) {
	expectedResponse := "file 1\n file 2\n file 3"

	listFilesEndpoint := func(writer http.ResponseWriter, request *http.Request) {
		methodMatches := request.Method == "GET"
		pathMatches := request.URL.Path == "/some/path"

		if !methodMatches || !pathMatches {
			fmt.Printf("One of the matchers did not match. Method [%t] Path [%t]",
				methodMatches, pathMatches)

			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
		fmt.Fprint(writer, expectedResponse)
	}

	listFilesServer := httptest.NewTLSServer(http.HandlerFunc(listFilesEndpoint))
	defer listFilesServer.Close()

	req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method: "GET",
		Path:   "/v2/apps/my-app-guid/instances/0/files/some/path",
		Response: testnet.TestResponse{
			Status: http.StatusTemporaryRedirect,
			Header: http.Header{
				"Location": {fmt.Sprintf("%s/some/path", listFilesServer.URL)},
			},
		},
	})

	listFilesRedirectServer, handler := testnet.NewTLSServer(t, []testnet.TestRequest{req})
	defer listFilesRedirectServer.Close()

	config := &configuration.Configuration{
		Target:      listFilesRedirectServer.URL,
		AccessToken: "BEARER my_access_token",
	}

	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerAppFilesRepository(config, gateway)
	list, err := repo.ListFiles("my-app-guid", "some/path")

	assert.True(t, handler.AllRequestsCalled())
	assert.False(t, err.IsNotSuccessful())
	assert.Equal(t, list, expectedResponse)
}
