package api_test

import (
	. "cf/api"
	"cf/net"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("AppFilesRepository", func() {
	It("lists files", func() {

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

		listFilesRedirectServer, handler := testnet.NewTLSServer([]testnet.TestRequest{req})
		defer listFilesRedirectServer.Close()

		configRepo := testconfig.NewRepositoryWithDefaults()
		configRepo.SetApiEndpoint(listFilesRedirectServer.URL)

		gateway := net.NewCloudControllerGateway()
		repo := NewCloudControllerAppFilesRepository(configRepo, gateway)
		list, err := repo.ListFiles("my-app-guid", "some/path")

		Expect(handler.AllRequestsCalled()).To(BeTrue())
		Expect(err.IsNotSuccessful()).To(BeFalse())
		Expect(list).To(Equal(expectedResponse))
	})
})
