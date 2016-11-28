package password_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api/password"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudControllerPasswordRepository", func() {
	It("updates your password", func() {
		req := apifakes.NewCloudControllerTestRequest(testnet.TestRequest{
			Method:   "PUT",
			Path:     "/Users/my-user-guid/password",
			Matcher:  testnet.RequestBodyMatcher(`{"password":"new-password","oldPassword":"old-password"}`),
			Response: testnet.TestResponse{Status: http.StatusOK},
		})

		passwordUpdateServer, handler, repo := createPasswordRepo(req)
		defer passwordUpdateServer.Close()

		apiErr := repo.UpdatePassword("old-password", "new-password")
		Expect(handler).To(HaveAllRequestsCalled())
		Expect(apiErr).NotTo(HaveOccurred())
	})
})

func createPasswordRepo(req testnet.TestRequest) (passwordServer *httptest.Server, handler *testnet.TestHandler, repo Repository) {
	passwordServer, handler = testnet.NewServer([]testnet.TestRequest{req})

	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetUaaEndpoint(passwordServer.URL)
	gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
	repo = NewCloudControllerRepository(configRepo, gateway)
	return
}
