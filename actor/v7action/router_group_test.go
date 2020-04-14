package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Router Group Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		warnings                  Warnings
		executeErr                error
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil)
	})

	Describe("GetRouterGroups()", func() {
		var (
			routerGroups []RouterGroup
		)

		When("the request succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRouterGroupsReturns(
					[]ccv3.RouterGroup{
						{Name: "router-group-name-1"},
						{Name: "router-group-name-2"},
					},
					ccv3.Warnings{"router-group-warning"},
					nil,
				)
			})

			JustBeforeEach(func() {
				routerGroups, warnings, executeErr = actor.GetRouterGroups()
				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("returns the security groups and warnings", func() {
				Expect(routerGroups).To(Equal(
					[]RouterGroup{
						{Name: "router-group-name-1"},
						{Name: "router-group-name-2"},
					},
				))

				Expect(warnings).To(ConsistOf("router-group-warning"))

				Expect(fakeCloudControllerClient.GetRouterGroupsCallCount()).To(Equal(1))
			})
		})

		When("the request errors", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetRouterGroupsReturns(
					nil,
					ccv3.Warnings{"router-group-warning"},
					expectedError,
				)
			})

			JustBeforeEach(func() {
				routerGroups, warnings, executeErr = actor.GetRouterGroups()
			})

			It("returns the error and warnings", func() {
				Expect(warnings).To(ConsistOf("router-group-warning"))
				Expect(executeErr).To(MatchError(expectedError))
			})
		})
	})
})
