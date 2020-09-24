package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance", func() {
	var (
		requester *ccv3fakes.FakeRequester
		client    *Client
	)

	BeforeEach(func() {
		requester = new(ccv3fakes.FakeRequester)
		client, _ = NewFakeRequesterTestClient(requester)
	})

	Describe("GetServiceCredentialBindings", func() {
		var (
			query        []Query
			bindings     []resources.ServiceCredentialBinding
			includedApps []resources.Application
			warnings     Warnings
			executeErr   error
		)

		BeforeEach(func() {
			requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
				types := []resources.ServiceCredentialBindingType{resources.KeyBinding, resources.AppBinding}
				for i := 1; i <= 3; i++ {
					Expect(requestParams.AppendToList(resources.ServiceCredentialBinding{
						GUID:                fmt.Sprintf("credential-binding-%d-guid", i),
						Name:                fmt.Sprintf("credential-binding-%d-name", i),
						ServiceInstanceGUID: fmt.Sprintf("si-%d-guid", i),
						AppGUID:             fmt.Sprintf("app-%d-guid", i),
						Type:                types[i%2],
					})).NotTo(HaveOccurred())
				}
				return IncludedResources{Apps: includedApps}, Warnings{"warning-1", "warning-2"}, nil
			})

			includedApps = nil

			query = []Query{
				{Key: ServiceInstanceGUIDFilter, Values: []string{"si-1-guid", "si-2-guid", "si-3-guid", "si-4-guid"}},
			}
		})

		JustBeforeEach(func() {
			bindings, warnings, executeErr = client.GetServiceCredentialBindings(query...)
		})

		It("makes the correct call", func() {
			Expect(requester.MakeListRequestCallCount()).To(Equal(1))
			actualParams := requester.MakeListRequestArgsForCall(0)
			Expect(actualParams.RequestName).To(Equal(internal.GetServiceCredentialBindingsRequest))
			Expect(actualParams.Query).To(ConsistOf(Query{Key: ServiceInstanceGUIDFilter, Values: []string{"si-1-guid", "si-2-guid", "si-3-guid", "si-4-guid"}}))
			Expect(actualParams.ResponseBody).To(BeAssignableToTypeOf(resources.ServiceCredentialBinding{}))
		})

		It("returns a list of service credential bindings", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

			Expect(bindings).To(ConsistOf(
				resources.ServiceCredentialBinding{
					GUID:                "credential-binding-1-guid",
					Name:                "credential-binding-1-name",
					ServiceInstanceGUID: "si-1-guid",
					AppGUID:             "app-1-guid",
					Type:                resources.AppBinding,
				},
				resources.ServiceCredentialBinding{
					GUID:                "credential-binding-2-guid",
					Name:                "credential-binding-2-name",
					ServiceInstanceGUID: "si-2-guid",
					AppGUID:             "app-2-guid",
					Type:                resources.KeyBinding,
				},
				resources.ServiceCredentialBinding{
					GUID:                "credential-binding-3-guid",
					Name:                "credential-binding-3-name",
					ServiceInstanceGUID: "si-3-guid",
					AppGUID:             "app-3-guid",
					Type:                resources.AppBinding,
				},
			))
		})

		When("app resources are included via the query", func() {
			BeforeEach(func() {
				query = append(query, Query{Key: Include, Values: []string{"app"}})

				includedApps = []resources.Application{
					{GUID: "app-1-guid", Name: "app-1", SpaceGUID: "space-1-guid"},
					{GUID: "app-3-guid", Name: "app-3", SpaceGUID: "space-2-guid"},
				}
			})

			It("returns the app names", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(bindings[0].AppName).To(Equal("app-1"))
				Expect(bindings[0].AppSpaceGUID).To(Equal("space-1-guid"))
				Expect(bindings[1].AppName).To(BeEmpty())
				Expect(bindings[1].AppSpaceGUID).To(BeEmpty())
				Expect(bindings[2].AppName).To(Equal("app-3"))
				Expect(bindings[2].AppSpaceGUID).To(Equal("space-2-guid"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				errors := []ccerror.V3Error{
					{
						Code:   42424,
						Detail: "Some detailed error message",
						Title:  "CF-SomeErrorTitle",
					},
					{
						Code:   11111,
						Detail: "Some other detailed error message",
						Title:  "CF-SomeOtherErrorTitle",
					},
				}

				requester.MakeListRequestReturns(
					IncludedResources{},
					Warnings{"this is a warning"},
					ccerror.MultiError{ResponseCode: http.StatusTeapot, Errors: errors},
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   42424,
							Detail: "Some detailed error message",
							Title:  "CF-SomeErrorTitle",
						},
						{
							Code:   11111,
							Detail: "Some other detailed error message",
							Title:  "CF-SomeOtherErrorTitle",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
