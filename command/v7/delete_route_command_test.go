package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-route Command", func() {
	var (
		cmd             DeleteRouteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeDeleteRouteActor
		input           *Buffer
		binaryName      string
		executeErr      error
		domain          string
		hostname        string
		path            string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeDeleteRouteActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		domain = "some-domain.com"
		hostname = "host"
		path = `path`

		cmd = DeleteRouteCommand{
			RequiredArgs: flag.Domain{Domain: domain},
			Hostname:     hostname,
			Path:         path,
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: "some-org-guid",
		})

		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("the -f flag is NOT provided", func() {
		BeforeEach(func() {
			cmd.Force = false
		})

		When("the user inputs yes", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("y\n"))
				Expect(err).ToNot(HaveOccurred())

				fakeActor.DeleteRouteReturns(v7action.Warnings{"some-warning"}, nil)
			})

			It("delegates to the Actor", func() {
				actualDomainName, actualHostname, actualPath := fakeActor.DeleteRouteArgsForCall(0)
				Expect(actualDomainName).To(Equal(domain))
				Expect(actualHostname).To(Equal(hostname))
				Expect(actualPath).To(Equal(path))
			})

			It("deletes the route", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Deleting route %s.%s/%s\.\.\.`, hostname, domain, path))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).NotTo(Say("Route 'some-domain' does not exist"))
			})
		})

		When("the user inputs no", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("cancels the delete", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("'%s.%s/%s' has not been deleted.", hostname, domain, path))
				Expect(fakeActor.DeleteRouteCallCount()).To(Equal(0))
			})
		})

		When("the user chooses the default", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("cancels the delete", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("'%s.%s/%s' has not been deleted.", hostname, domain, path))
				Expect(fakeActor.DeleteRouteCallCount()).To(Equal(0))
			})
		})

		When("the user input is invalid", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("e\n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("asks the user again", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say(`Really delete the route host.some-domain\.com/path\? \[yN\]`))
				Expect(testUI.Out).To(Say(`invalid input \(not y, n, yes, or no\)`))
				Expect(testUI.Out).To(Say(`Really delete the route host.some-domain\.com/path\? \[yN\]`))

				Expect(fakeActor.DeleteRouteCallCount()).To(Equal(0))
			})
		})
	})

	When("the -f flag is provided", func() {
		BeforeEach(func() {
			cmd.Force = true
		})

		When("deleting the route errors", func() {
			Context("generic error", func() {
				BeforeEach(func() {
					fakeActor.DeleteRouteReturns(v7action.Warnings{"some-warning"}, errors.New("some-error"))
				})

				It("displays all warnings, and returns the error", func() {
					Expect(testUI.Err).To(Say("some-warning"))
					Expect(testUI.Out).To(Say(`Deleting route host.some-domain.com\/path\.\.\.`))
					Expect(testUI.Out).ToNot(Say("OK"))
					Expect(executeErr).To(MatchError("some-error"))
				})
			})
		})

		When("the route doesn't exist", func() {
			BeforeEach(func() {
				fakeActor.DeleteRouteReturns(
					v7action.Warnings{"some-warning"},
					actionerror.RouteNotFoundError{
						Host:       hostname,
						DomainName: domain,
						Path:       path},
				)
			})

			It("displays all warnings, that the domain wasn't found, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Deleting route %s.%s/%s\.\.\.`, hostname, domain, path))
				Expect(testUI.Out).To(Say(`Unable to delete\. Route with host '%s', domain '%s', and path '%s' not found.`, hostname, domain, path))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("the route exists", func() {
			BeforeEach(func() {
				fakeActor.DeleteRouteReturns(v7action.Warnings{"some-warning"}, nil)
			})

			It("displays all warnings, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Deleting route %s\.%s/%s\.\.\.`, hostname, domain, path))
				Expect(testUI.Out).To(Say("OK"))
			})
		})
	})
})
