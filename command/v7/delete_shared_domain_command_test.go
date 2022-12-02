package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-shared-domain Command", func() {
	var (
		cmd             DeleteSharedDomainCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
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
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		domain = "some-domain.com"

		cmd = DeleteSharedDomainCommand{
			RequiredArgs: flag.Domain{Domain: domain},

			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: "some-org-guid",
		})

		fakeActor.GetCurrentUserReturns(configv3.User{Name: "steve"}, nil)
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
			fakeActor.GetCurrentUserReturns(configv3.User{}, expectedErr)
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

				fakeActor.GetDomainByNameReturns(resources.Domain{Name: "some-domain.com", GUID: "some-guid"}, v7action.Warnings{"some-warning1"}, nil)

				fakeActor.DeleteDomainReturns(v7action.Warnings{"some-warning2"}, nil)
			})

			It("delegates to the Actor", func() {
				domain := fakeActor.DeleteDomainArgsForCall(0)
				Expect(domain).To(Equal(resources.Domain{Name: "some-domain.com", GUID: "some-guid"}))
			})

			It("deletes the shared domain", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning1"))
				Expect(testUI.Err).To(Say("some-warning2"))
				Expect(testUI.Out).To(Say(`Deleting domain some-domain.com as steve\.\.\.`))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).NotTo(Say("Domain some-domain does not exist"))
			})

			When("GetDomainByName() errors", func() {
				BeforeEach(func() {
					fakeActor.GetDomainByNameReturns(resources.Domain{Name: "some-domain.com", GUID: "some-guid"}, v7action.Warnings{"some-warning"}, errors.New("get-domain-by-name-errors"))
				})

				It("returns an error", func() {
					Expect(fakeActor.GetDomainByNameCallCount()).To(Equal(1))
					actualDomainName := fakeActor.GetDomainByNameArgsForCall(0)
					Expect(actualDomainName).To(Equal(domain))

					Expect(fakeActor.DeleteDomainCallCount()).To(Equal(0))

					Expect(testUI.Err).To(Say("some-warning"))
					Expect(executeErr).To(MatchError(errors.New("get-domain-by-name-errors")))
				})
			})

			When("the domain is private, not shared", func() {
				BeforeEach(func() {
					fakeActor.GetDomainByNameReturns(resources.Domain{Name: "some-domain.com", GUID: "some-guid", OrganizationGUID: "private-org-guid"}, v7action.Warnings{"some-warning"}, nil)
				})
				It("returns an informative error message and does not delete the domain", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr).To(MatchError(translatableerror.NotSharedDomainError{DomainName: "some-domain.com"}))
					Expect(testUI.Out).To(Say(`Deleting domain some-domain.com as steve\.\.\.`))
					Expect(testUI.Err).To(Say("some-warning"))
					Expect(fakeActor.DeleteDomainCallCount()).To(Equal(0))
				})
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
				Expect(fakeActor.DeleteDomainCallCount()).To(Equal(0))
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
				Expect(fakeActor.DeleteDomainCallCount()).To(Equal(0))
			})
		})

		When("the user input is invalid", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("e\n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("asks the user again", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say(`Really delete the shared domain some-domain.com\? \[yN\]`))
				Expect(testUI.Out).To(Say(`invalid input \(not y, n, yes, or no\)`))
				Expect(testUI.Out).To(Say(`Really delete the shared domain some-domain.com\? \[yN\]`))

				Expect(fakeActor.DeleteDomainCallCount()).To(Equal(0))
			})
		})
	})

	When("the -f flag is provided", func() {
		BeforeEach(func() {
			cmd.Force = true
		})

		When("deleting the shared domain errors", func() {
			Context("generic error", func() {
				BeforeEach(func() {
					fakeActor.DeleteDomainReturns(v7action.Warnings{"some-warning"}, errors.New("some-error"))
				})

				It("displays all warnings, and returns the error", func() {
					Expect(testUI.Err).To(Say("some-warning"))
					Expect(testUI.Out).To(Say(`Deleting domain some-domain.com as steve\.\.\.`))
					Expect(testUI.Out).ToNot(Say("OK"))
					Expect(executeErr).To(MatchError("some-error"))
				})
			})
		})

		When("the shared domain doesn't exist", func() {
			BeforeEach(func() {
				fakeActor.GetDomainByNameReturns(resources.Domain{}, v7action.Warnings{"some-warning"}, actionerror.DomainNotFoundError{Name: "some-domain.com"})
			})

			It("displays all warnings, that the domain wasn't found, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Deleting domain some-domain.com as steve\.\.\.`))
				Expect(testUI.Err).To(Say(`Domain 'some-domain\.com' does not exist\.`))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("the shared domain exists", func() {
			BeforeEach(func() {
				fakeActor.DeleteDomainReturns(v7action.Warnings{"some-warning"}, nil)
			})

			It("displays all warnings, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Deleting domain some-domain.com as steve\.\.\.`))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Err).NotTo(Say(`Domain 'some-domain\.com' does not exist\.`))
			})
		})
	})
})
