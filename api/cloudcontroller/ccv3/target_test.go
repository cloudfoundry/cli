package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Target", func() {
	var (
		client    *Client
		serverURL string
	)

	BeforeEach(func() {
		client = NewClient(Config{AppName: "CF CLI API V3 Target Test", AppVersion: "Unknown"})
	})

	Describe("TargetCF", func() {
		BeforeEach(func() {
			server.Reset()

			serverURL = server.URL()
			rootResponse := fmt.Sprintf(`{
				"links": {
					"self": {
						"href": "%s"
					},
					"cloud_controller_v2": {
						"href": "%s/v2",
						"meta": {
							"version": "2.64.0"
						}
					},
					"cloud_controller_v3": {
						"href": "%s/v3",
						"meta": {
							"version": "3.0.0-alpha.5"
						}
					},
					"uaa": {
						"href": "https://uaa.bosh-lite.com"
					}
				}
			}`, serverURL, serverURL, serverURL)

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/"),
					RespondWith(
						http.StatusOK,
						rootResponse,
						http.Header{"X-Cf-Warnings": {"warning 1"}}),
				),
			)
		})

		When("client has wrappers", func() {
			var fakeWrapper1 *ccv3fakes.FakeConnectionWrapper
			var fakeWrapper2 *ccv3fakes.FakeConnectionWrapper

			BeforeEach(func() {
				fakeWrapper1 = new(ccv3fakes.FakeConnectionWrapper)
				fakeWrapper1.WrapReturns(fakeWrapper1)
				fakeWrapper2 = new(ccv3fakes.FakeConnectionWrapper)
				fakeWrapper2.WrapReturns(fakeWrapper2)

				fakeWrapper2.MakeStub = func(request *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
					apiInfo, ok := passedResponse.DecodeJSONResponseInto.(*Info)
					if ok { // Only caring about the first time Make is called, ignore all others
						apiInfo.Links.CCV3.HREF = server.URL() + "/v3"
					}
					return nil
				}

				client = NewClient(Config{
					AppName:    "CF CLI API Target Test",
					AppVersion: "Unknown",
					Wrappers:   []ConnectionWrapper{fakeWrapper1, fakeWrapper2},
				})
			})

			It("calls wrap on all the wrappers", func() {
				client.TargetCF(TargetSettings{
					SkipSSLValidation: true,
					URL:               server.URL(),
				})

				Expect(fakeWrapper1.WrapCallCount()).To(Equal(1))
				Expect(fakeWrapper2.WrapCallCount()).To(Equal(1))
				Expect(fakeWrapper2.WrapArgsForCall(0)).To(Equal(fakeWrapper1))
			})
		})
	})
})
