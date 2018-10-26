package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/router"
	"code.cloudfoundry.org/cli/api/router/routererror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Router Group Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		fakeRouterClient          *v2actionfakes.FakeRouterClient
	)

	BeforeEach(func() {
		fakeRouterClient = new(v2actionfakes.FakeRouterClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetRouterGroupByName", func() {
		var (
			routerGroupName string
			routerGroup     RouterGroup
			err             error
		)

		JustBeforeEach(func() {
			routerGroup, err = actor.GetRouterGroupByName(routerGroupName, fakeRouterClient)
		})

		When("the router group does not exist", func() {
			BeforeEach(func() {
				routerGroupName = "some-router-group"
				fakeRouterClient.GetRouterGroupsByNameReturns([]router.RouterGroup{}, routererror.ErrorResponse{
					Message:    "Router Group 'some-router-group' not found",
					StatusCode: 404,
					Name:       "ResourceNotFoundError",
				})
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(actionerror.RouterGroupNotFoundError{Name: routerGroupName}))
				Expect(routerGroup).To(Equal(RouterGroup{}))
				Expect(fakeRouterClient.GetRouterGroupsByNameCallCount()).To(Equal(1))
			})
		})

		When("the router group exists", func() {
			BeforeEach(func() {
				routerGroupName = "default-tcp"
				fakeRouterClient.GetRouterGroupsByNameReturns([]router.RouterGroup{router.RouterGroup{Name: routerGroupName}}, nil)
			})

			It("should return the router group and not an error", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(routerGroup).To(Equal(RouterGroup{Name: routerGroupName}))
				Expect(fakeRouterClient.GetRouterGroupsByNameCallCount()).To(Equal(1))
			})
		})

		When("the router client returns an error", func() {
			BeforeEach(func() {
				routerGroupName = "default-tcp"
				fakeRouterClient.GetRouterGroupsByNameReturns([]router.RouterGroup{}, errors.New("The request failed"))
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("The request failed"))
				Expect(fakeRouterClient.GetRouterGroupsByNameCallCount()).To(Equal(1))
			})
		})
	})
})
