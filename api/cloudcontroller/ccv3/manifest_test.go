package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Manifest", func() {
	var (
		client    *Client
		requester *ccv3fakes.FakeRequester
	)

	BeforeEach(func() {
		requester = new(ccv3fakes.FakeRequester)
		client, _ = NewFakeRequesterTestClient(requester)
	})

	Describe("GetApplicationManifest", func() {
		var (
			appGUID string

			rawManifest []byte
			warnings    Warnings
			executeErr  error

			expectedYAML []byte
		)

		BeforeEach(func() {
			appGUID = "some-app-guid"
		})

		JustBeforeEach(func() {
			rawManifest, warnings, executeErr = client.GetApplicationManifest(appGUID)
		})

		When("getting the manifest is successful", func() {
			BeforeEach(func() {
				expectedYAML = []byte("---\n- banana")
				requester.MakeRequestReceiveRawReturns(expectedYAML, Warnings{"this is a warning"}, nil)
			})

			It("makes the correct request", func() {
				Expect(requester.MakeRequestReceiveRawCallCount()).To(Equal(1))
				requestName, uriParams, responseBody := requester.MakeRequestReceiveRawArgsForCall(0)
				Expect(requestName).To(Equal(internal.GetApplicationManifestRequest))
				Expect(uriParams).To(Equal(internal.Params{"app_guid": appGUID}))
				Expect(responseBody).To(Equal("application/x-yaml"))
			})

			It("the manifest and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(rawManifest).To(Equal(expectedYAML))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   10008,
						Detail: "The request is semantically invalid: command presence",
						Title:  "CF-UnprocessableEntity",
					},
					{
						Code:   10010,
						Detail: "App not found",
						Title:  "CF-AppNotFound",
					},
				}

				requester.MakeRequestReceiveRawReturns(
					nil,
					Warnings{"this is a warning"},
					ccerror.MultiError{ResponseCode: http.StatusTeapot, Errors: errors},
				)

			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "App not found",
							Title:  "CF-AppNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
