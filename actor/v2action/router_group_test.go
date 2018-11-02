package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/router"
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
			routerGroup RouterGroup
			executeErr  error
		)

		JustBeforeEach(func() {
			routerGroup, executeErr = actor.GetRouterGroupByName("some-router-group-name", fakeRouterClient)
		})

		When("the router group exists", func() {
			BeforeEach(func() {
				fakeRouterClient.GetRouterGroupByNameReturns(
					router.RouterGroup{Name: "some-router-group-name"},
					nil)
			})

			It("should return the router group and not an error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(routerGroup).To(Equal(RouterGroup{Name: "some-router-group-name"}))
				Expect(fakeRouterClient.GetRouterGroupByNameCallCount()).To(Equal(1))
			})
		})

		When("the router client returns an error", func() {
			BeforeEach(func() {
				fakeRouterClient.GetRouterGroupByNameReturns(router.RouterGroup{}, errors.New("The request failed"))
			})

			It("should return an error", func() {
				Expect(executeErr).To(MatchError("The request failed"))
				Expect(fakeRouterClient.GetRouterGroupByNameCallCount()).To(Equal(1))
			})
		})
	})
})
