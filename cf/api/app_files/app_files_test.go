package app_files_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"

	. "github.com/cloudfoundry/cli/cf/api/app_files"
	"github.com/cloudfoundry/cli/testhelpers/cloud_controller_gateway"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

		listFilesServer := httptest.NewServer(http.HandlerFunc(listFilesEndpoint))
		defer listFilesServer.Close()

		req := testapi.NewCloudControllerTestRequest(testnet.TestRequest{
			Method: "GET",
			Path:   "/v2/apps/my-app-guid/instances/1/files/some/path",
			Response: testnet.TestResponse{
				Status: http.StatusTemporaryRedirect,
				Header: http.Header{
					"Location": {fmt.Sprintf("%s/some/path", listFilesServer.URL)},
				},
			},
		})

		listFilesRedirectServer, handler := testnet.NewServer([]testnet.TestRequest{req})
		defer listFilesRedirectServer.Close()

		configRepo := testconfig.NewRepositoryWithDefaults()
		configRepo.SetApiEndpoint(listFilesRedirectServer.URL)

		gateway := cloud_controller_gateway.NewTestCloudControllerGateway(configRepo)
		repo := NewCloudControllerAppFilesRepository(configRepo, gateway)
		list, err := repo.ListFiles("my-app-guid", 1, "some/path")

		Expect(handler).To(HaveAllRequestsCalled())
		Expect(err).ToNot(HaveOccurred())
		Expect(list).To(Equal(expectedResponse))
	})
})
