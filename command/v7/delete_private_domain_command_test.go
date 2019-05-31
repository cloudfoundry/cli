package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"

	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
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

var _ = Describe("delete-private-domain Command", func() {
	var (
		cmd             DeletePrivateDomainCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeDeletePrivateDomainActor
		input           *Buffer
		binaryName      string
		executeErr      error
		domain          string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeDeletePrivateDomainActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		domain = "some-domain.com"

		cmd = DeletePrivateDomainCommand{
			RequiredArgs: flag.Domain{Domain: domain},

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
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

	When("the user attempts to delete-private-domain a shared domain", func() {
		BeforeEach(func() {
			fakeActor.CheckSharedDomainReturns(v7action.Warnings{"some-warning"}, fmt.Errorf(`Domain %s is a shared domain, not a private domain.`, domain))
		})
	})

	When("the domain does not exist", func() {
		BeforeEach(func() {
			fakeActor.CheckSharedDomainReturns(nil, actionerror.DomainNotFoundError{Name: "domain.com"})
		})

		It("displays OK and returns with success", func() {
			Expect(testUI.Out).To(Say("Domain some-domain.com does not exist"))
			Expect(testUI.Out).To(Say("OK"))
			Expect(executeErr).ToNot(HaveOccurred())
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

				fakeActor.DeletePrivateDomainReturns(v7action.Warnings{"some-warning"}, nil)
			})

			It("delegates to the Actor", func() {
				actualName := fakeActor.DeletePrivateDomainArgsForCall(0)
				Expect(actualName).To(Equal(domain))
			})

			It("deletes the private domain", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Deleting private domain some-domain.com as steve\.\.\.`))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).NotTo(Say("Domain some-domain does not exist"))
			})
		})

		When("the user inputs no", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("cancels the delete", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("'some-domain.com' has not been deleted."))
				Expect(fakeActor.DeletePrivateDomainCallCount()).To(Equal(0))
			})
		})

		When("the user chooses the default", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("cancels the delete", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say(`\'some-domain.com\' has not been deleted.`))
				Expect(fakeActor.DeletePrivateDomainCallCount()).To(Equal(0))
			})
		})

		When("the user input is invalid", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("e\n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("asks the user again", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say(`Really delete the private domain some-domain.com\? \[yN\]`))
				Expect(testUI.Out).To(Say(`invalid input \(not y, n, yes, or no\)`))
				Expect(testUI.Out).To(Say(`Really delete the private domain some-domain.com\? \[yN\]`))

				Expect(fakeActor.DeletePrivateDomainCallCount()).To(Equal(0))
			})
		})
	})

	When("the -f flag is provided", func() {
		BeforeEach(func() {
			cmd.Force = true
		})

		When("deleting the private domain errors", func() {
			Context("generic error", func() {
				BeforeEach(func() {
					fakeActor.DeletePrivateDomainReturns(v7action.Warnings{"some-warning"}, errors.New("some-error"))
				})

				It("displays all warnings, and returns the error", func() {
					Expect(testUI.Err).To(Say("some-warning"))
					Expect(testUI.Out).To(Say(`Deleting private domain some-domain.com as steve\.\.\.`))
					Expect(testUI.Out).ToNot(Say("OK"))
					Expect(executeErr).To(MatchError("some-error"))
				})
			})
		})

		When("the private domain doesn't exist", func() {
			BeforeEach(func() {
				fakeActor.DeletePrivateDomainReturns(v7action.Warnings{"some-warning"}, actionerror.DomainNotFoundError{Name: "some-domain.com"})
			})

			It("displays all warnings, that the domain wasn't found, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Deleting private domain some-domain.com as steve\.\.\.`))
				Expect(testUI.Out).To(Say("Domain 'some-domain.com' not found."))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("the private domain exists", func() {
			BeforeEach(func() {
				fakeActor.DeletePrivateDomainReturns(v7action.Warnings{"some-warning"}, nil)
			})

			It("displays all warnings, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Deleting private domain some-domain.com as steve\.\.\.`))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).NotTo(Say("Domain some-domain.com does not exist"))
			})
		})
	})
})
