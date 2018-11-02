package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"

	"code.cloudfoundry.org/cli/actor/v2action"

	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("CreateSharedDomainCommand", func() {
	var (
		fakeConfig      *commandfakes.FakeConfig
		fakeActor       *v6fakes.FakeCreateSharedDomainActor
		fakeSharedActor *commandfakes.FakeSharedActor
		testUI          *ui.UI
		cmd             CreateSharedDomainCommand

		executeErr       error
		sharedDomainName string
		username         string
		routerGroupName  string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(v6fakes.FakeCreateSharedDomainActor)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		sharedDomainName = "some-shared-domain-name"
	})

	JustBeforeEach(func() {
		cmd = CreateSharedDomainCommand{
			UI:           testUI,
			Config:       fakeConfig,
			Actor:        fakeActor,
			SharedActor:  fakeSharedActor,
			RequiredArgs: flag.Domain{Domain: sharedDomainName},
			RouterGroup:  routerGroupName,
		}

		executeErr = cmd.Execute(nil)
	})

	It("checks for user being logged in", func() {
		Expect(fakeSharedActor.RequireCurrentUserCallCount()).To(Equal(1))
	})

	When("user is logged in", func() {
		BeforeEach(func() {
			username = "some-user-name"
			fakeSharedActor.RequireCurrentUserReturns(username, nil)
		})

		When("the user is logged in as an admin", func() {
			When("--router-group is passed", func() {
				When("the router group does not exists", func() {
					var actorError error
					BeforeEach(func() {
						routerGroupName = "some-router-group"
						actorError = actionerror.RouterGroupNotFoundError{Name: routerGroupName}
						fakeActor.GetRouterGroupByNameReturns(v2action.RouterGroup{}, actorError)
					})

					It("should fail and return a translateable error", func() {
						Expect(testUI.Out).To(Say("Creating shared domain %s as %s...", sharedDomainName, username))
						namePassed, _ := fakeActor.GetRouterGroupByNameArgsForCall(0)
						Expect(namePassed).To(Equal(routerGroupName))
						Expect(executeErr).To(MatchError(actorError))
					})
				})

				When("the router group is found", func() {
					var routerGroupGUID string

					BeforeEach(func() {
						routerGroupName = "some-router-group"
						routerGroupGUID = "some-guid"
						fakeActor.GetRouterGroupByNameReturns(v2action.RouterGroup{
							Name: routerGroupName,
							GUID: routerGroupGUID,
						}, nil)
					})

					It("should create the domain with the router group", func() {
						domainName, routerGroup := fakeActor.CreateSharedDomainArgsForCall(0)
						Expect(domainName).To(Equal(sharedDomainName))
						Expect(routerGroup).To(Equal(v2action.RouterGroup{
							Name: routerGroupName,
							GUID: routerGroupGUID,
						}))
					})
				})
			})

			When("--router-group is not passed", func() {
				BeforeEach(func() {
					routerGroupName = ""
				})

				It("does not call fetch the router group", func() {
					Expect(fakeActor.GetRouterGroupByNameCallCount()).To(Equal(0))
				})

				It("attempts to create the shared domain", func() {
					domainNamePassed, routerGroup := fakeActor.CreateSharedDomainArgsForCall(0)
					Expect(domainNamePassed).To(Equal(cmd.RequiredArgs.Domain))
					Expect(routerGroup).To(Equal(v2action.RouterGroup{}))
				})
			})
		})

		When("the user is not logged in as an admin", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("must be admin")
				fakeActor.CreateSharedDomainReturns(v2action.Warnings{"warning-1", "warning-2"}, expectedError)
			})

			It("returns the unauthorized error from CC API", func() {
				Expect(fakeActor.CreateSharedDomainCallCount()).To(Equal(1))
				Expect(executeErr).To(MatchError(expectedError))
			})
		})
	})

	When("the user is not logger in", func() {
		expectedErr := errors.New("not logged in and/or can't verify login because of error")

		BeforeEach(func() {
			fakeSharedActor.RequireCurrentUserReturns("", expectedErr)
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError(expectedErr))
		})
	})

})
