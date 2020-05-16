package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/router"
	"code.cloudfoundry.org/cli/api/router/routererror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Router Group Actions", func() {
	var (
		actor             *Actor
		fakeRoutingClient *v7actionfakes.FakeRoutingClient
		executeErr        error
	)

	BeforeEach(func() {
		fakeRoutingClient = new(v7actionfakes.FakeRoutingClient)
		actor = NewActor(nil, nil, nil, nil, fakeRoutingClient, nil)
	})

	Describe("GetRouterGroups", func() {
		var (
			routerGroups []RouterGroup
		)

		When("the request succeeds", func() {
			BeforeEach(func() {
				fakeRoutingClient.GetRouterGroupsReturns(
					[]router.RouterGroup{
						{Name: "router-group-name-1"},
						{Name: "router-group-name-2"},
					},
					nil,
				)
			})

			JustBeforeEach(func() {
				routerGroups, executeErr = actor.GetRouterGroups()
				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("returns the router groups", func() {
				Expect(routerGroups).To(Equal(
					[]RouterGroup{
						{Name: "router-group-name-1"},
						{Name: "router-group-name-2"},
					},
				))

				Expect(fakeRoutingClient.GetRouterGroupsCallCount()).To(Equal(1))
			})
		})

		When("the request errors", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a Routing API Error")

				fakeRoutingClient.GetRouterGroupsReturns(
					nil,
					expectedError,
				)
			})

			JustBeforeEach(func() {
				routerGroups, executeErr = actor.GetRouterGroups()
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedError))
			})
		})
	})

	Describe("GetRouterGroupByName", func() {
		var (
			routerGroup RouterGroup
			name        string
		)

		When("the request succeeds", func() {
			BeforeEach(func() {
				name = "router-group-name-1"

				fakeRoutingClient.GetRouterGroupByNameReturns(
					router.RouterGroup{Name: name},
					nil,
				)
			})

			JustBeforeEach(func() {
				routerGroup, executeErr = actor.GetRouterGroupByName(name)
				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("returns the router groups", func() {
				Expect(routerGroup).To(Equal(
					RouterGroup{
						Name: "router-group-name-1",
					},
				))

				Expect(fakeRoutingClient.GetRouterGroupByNameCallCount()).To(Equal(1))
				givenName := fakeRoutingClient.GetRouterGroupByNameArgsForCall(0)
				Expect(givenName).To(Equal(name))
			})
		})

		When("there are no router groups found", func() {
			BeforeEach(func() {
				name = "router-group-name-1"

				fakeRoutingClient.GetRouterGroupByNameReturns(
					router.RouterGroup{},
					routererror.ResourceNotFoundError{},
				)
			})

			JustBeforeEach(func() {
				routerGroup, executeErr = actor.GetRouterGroupByName(name)
			})

			It("returns a helpful error", func() {
				Expect(routerGroup).To(Equal(RouterGroup{}))
				Expect(executeErr).To(MatchError(actionerror.RouterGroupNotFoundError{Name: name}))
			})
		})

		When("the request errors", func() {
			var expectedError error

			BeforeEach(func() {
				name = "router-group-name-1"
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeRoutingClient.GetRouterGroupByNameReturns(
					router.RouterGroup{},
					expectedError,
				)
			})

			JustBeforeEach(func() {
				routerGroup, executeErr = actor.GetRouterGroupByName(name)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedError))
			})
		})
	})
})
