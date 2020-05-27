package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-security-group Command", func() {
	var (
		cmd             DeleteSecurityGroupCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		input           *Buffer
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = DeleteSecurityGroupCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		cmd.RequiredArgs.SecurityGroup = "some-security-group"
		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("a cloud controller API endpoint is set", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-url")
		})

		When("checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeFalse())
				Expect(checkTargetedSpace).To(BeFalse())
			})
		})

		When("the user is logged in", func() {
			When("getting the current user returns an error", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = errors.New("some error")
					fakeConfig.CurrentUserReturns(configv3.User{}, returnedErr)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError(returnedErr))
				})
			})

			When("getting the current user does not return an error", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(
						configv3.User{Name: "some-user"},
						nil)
				})

				When("the '-f' flag is provided", func() {
					BeforeEach(func() {
						cmd.Force = true
					})

					When("no errors are encountered", func() {
						BeforeEach(func() {
							fakeActor.DeleteSecurityGroupReturns(v7action.Warnings{"warning-1", "warning-2"}, nil)
						})

						It("does not prompt for user confirmation, displays warnings, and deletes the security group", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).ToNot(Say("Really delete the security group some-security-group"))
							Expect(testUI.Out).To(Say("Deleting security group some-security-group as some-user..."))

							Expect(fakeActor.DeleteSecurityGroupCallCount()).To(Equal(1))
							securityGroupName := fakeActor.DeleteSecurityGroupArgsForCall(0)
							Expect(securityGroupName).To(Equal("some-security-group"))

							Expect(testUI.Err).To(Say("warning-1"))
							Expect(testUI.Err).To(Say("warning-2"))
							Expect(testUI.Out).To(Say("OK"))
						})
					})

					When("an error is encountered deleting the security group", func() {
						BeforeEach(func() {
							fakeActor.DeleteSecurityGroupReturns(
								v7action.Warnings{"warning-1", "warning-2"},
								actionerror.SecurityGroupNotFoundError{
									Name: "some-security-group",
								},
							)
						})

						It("returns an SecurityGroupNotFoundError and displays all warnings", func() {
							Expect(executeErr).NotTo(HaveOccurred())

							Expect(testUI.Out).To(Say("Deleting security group some-security-group as some-user..."))

							Expect(fakeActor.DeleteSecurityGroupCallCount()).To(Equal(1))
							securityGroupName := fakeActor.DeleteSecurityGroupArgsForCall(0)
							Expect(securityGroupName).To(Equal("some-security-group"))

							Expect(testUI.Err).To(Say("warning-1"))
							Expect(testUI.Err).To(Say("warning-2"))

							Expect(testUI.Err).To(Say(`Security group 'some-security-group' does not exist\.`))
							Expect(testUI.Out).To(Say("OK"))
						})
					})
				})

				// Testing the prompt.
				When("the '-f' flag is not provided", func() {
					When("the user chooses the default", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("does not delete the security group", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say(`Security group 'some-security-group' has not been deleted\.`))

							Expect(fakeActor.DeleteSecurityGroupCallCount()).To(Equal(0))
						})
					})

					When("the user inputs no", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("n\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("does not delete the security group", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say(`Security group 'some-security-group' has not been deleted\.`))

							Expect(fakeActor.DeleteSecurityGroupCallCount()).To(Equal(0))
						})
					})

					When("the user inputs yes", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("y\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("deletes the security group", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("Really delete the security group some-security-group"))
							Expect(testUI.Out).To(Say("Deleting security group some-security-group as some-user..."))

							Expect(fakeActor.DeleteSecurityGroupCallCount()).To(Equal(1))
							securityGroupName := fakeActor.DeleteSecurityGroupArgsForCall(0)
							Expect(securityGroupName).To(Equal("some-security-group"))

							Expect(testUI.Out).To(Say("OK"))
						})
					})

					When("the user input is invalid", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("e\n\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("does not delete the security group", func() {
							Expect(executeErr).NotTo(HaveOccurred())

							Expect(testUI.Out).To(Say("Really delete the security group some-security-group"))
							Expect(fakeActor.DeleteSecurityGroupCallCount()).To(Equal(0))
						})
					})

					When("displaying the prompt returns an error", func() {
						// if nothing is written to input, display bool prompt returns EOF
						It("returns the error", func() {
							Expect(executeErr).To(MatchError("EOF"))
						})
					})
				})
			})
		})
	})
})
