package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Info Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("GetRootResponse", func() {
		When("getting root is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRootReturns(
					ccv3.Root{
						Links: ccv3.RootLinks{
							LogCache: resources.APILink{HREF: "some-log-cache-url"},
						},
					},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns all warnings and root info", func() {
				rootInfo, warnings, err := actor.GetRootResponse()
				Expect(err).ToNot(HaveOccurred())

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetRootCallCount()).To(Equal(1))
				Expect(rootInfo.Links.LogCache.HREF).To(Equal("some-log-cache-url"))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetRootReturns(
					ccv3.Root{},
					ccv3.Warnings{"warning-1", "warning-2"},
					expectedErr,
				)
			})

			It("returns the same error and all warnings", func() {
				_, warnings, err := actor.GetRootResponse()
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GetInfoResponse", func() {
		When("getting info is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetInfoReturns(
					ccv3.Info{
						Name:          "test-name",
						Build:         "test-build",
						OSBAPIVersion: "1.0",
					},
					ccv3.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns all warnings and info", func() {
				info, warnings, err := actor.GetInfoResponse()
				Expect(err).ToNot(HaveOccurred())

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetInfoCallCount()).To(Equal(1))
				Expect(info.Name).To(Equal("test-name"))
				Expect(info.Build).To(Equal("test-build"))
				Expect(info.OSBAPIVersion).To(Equal("1.0"))
			})
		})

		When("the cloud controller client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetInfoReturns(
					ccv3.Info{},
					ccv3.Warnings{"warning-1", "warning-2"},
					expectedErr,
				)
			})

			It("returns the same error and all warnings", func() {
				_, warnings, err := actor.GetInfoResponse()
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})
})
