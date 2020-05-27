package v2action_test

import (
	"fmt"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/router"
	routererror "code.cloudfoundry.org/cli/api/router/routererror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Router Group Actions", func() {
	var (
		actor            *Actor
		fakeRouterClient *v2actionfakes.FakeRouterClient
	)

	BeforeEach(func() {
		fakeRouterClient = new(v2actionfakes.FakeRouterClient)
		actor = NewActor(nil, nil, nil)
	})

	Describe("GetRouterGroupByName", func() {
		var (
			routerGroupName string
			routerGroup     RouterGroup
			executeErr      error
		)

		JustBeforeEach(func() {
			routerGroup, executeErr = actor.GetRouterGroupByName(routerGroupName, fakeRouterClient)
		})

		When("the router group exists", func() {
			BeforeEach(func() {
				routerGroupName = "some-router-group-name"
				fakeRouterClient.GetRouterGroupByNameReturns(
					router.RouterGroup{Name: routerGroupName},
					nil)
			})

			It("should return the router group and not an error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(routerGroup).To(Equal(RouterGroup{Name: routerGroupName}))
				Expect(fakeRouterClient.GetRouterGroupByNameCallCount()).To(Equal(1))
			})
		})

		When("the router client returns an error", func() {
			When("it returns a ResourceNotFoundError", func() {
				BeforeEach(func() {
					routerGroupName = "tcp-default"
					fakeRouterClient.GetRouterGroupByNameReturns(router.RouterGroup{Name: routerGroupName}, routererror.ResourceNotFoundError{Message: "Router Group not found"})
				})

				It("should return an error", func() {
					Expect(executeErr).To(MatchError(fmt.Sprintf("Router group '%s' not found.", routerGroupName)))
					Expect(fakeRouterClient.GetRouterGroupByNameCallCount()).To(Equal(1))
				})
			})
			When("it returns a UnauthorizedError", func() {
				BeforeEach(func() {
					routerGroupName = "tcp-default"
					fakeRouterClient.GetRouterGroupByNameReturns(router.RouterGroup{Name: routerGroupName}, routererror.InvalidAuthTokenError{Message: "You are not authorized to peform the requested action"})
				})

				It("should return an error", func() {
					Expect(executeErr).To(MatchError("You are not authorized to peform the requested action"))
					Expect(fakeRouterClient.GetRouterGroupByNameCallCount()).To(Equal(1))
				})
			})
		})
	})
})
