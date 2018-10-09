package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-service-key Command", func() {
	var (
		cmd             CreateServiceKeyCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeCreateServiceKeyActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeCreateServiceKeyActor)

		cmd = CreateServiceKeyCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			RequiredArgs: flag.ServiceInstanceKey{
				ServiceInstance: "my-service",
				ServiceKey:      "my-key",
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns("faceman")
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
				Expect(checkTargetedOrg).To(BeTrue())
				Expect(checkTargetedSpace).To(BeTrue())
			})
		})

		When("the user is logged in, and an org and space are targeted", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
				fakeConfig.HasTargetedOrganizationReturns(true)
				fakeConfig.HasTargetedSpaceReturns(true)
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					GUID: "some-org-guid",
					Name: "some-org",
				})
				fakeConfig.TargetedSpaceReturns(configv3.Space{
					GUID: "some-space-guid",
					Name: "some-space",
				})
			})

			When("Sevice instance name and key name are passed", func() {
				When("warnings are being returned", func() {
					BeforeEach(func() {
						fakeActor.CreateServiceKeyReturns(v2action.ServiceKey{}, v2action.Warnings{"some-warning", "another-warning"}, nil)
					})

					It("displays the warning", func() {
						Expect(testUI.Err).To(Say("some-warning"))
						Expect(testUI.Err).To(Say("another-warning"))
					})
				})

				When("the service create is successful", func() {
					It("displays flavor text and creates the service key", func() {
						Expect(fakeActor.CreateServiceKeyCallCount()).To(Equal(1))
						service, key, spaceGUID, _ := fakeActor.CreateServiceKeyArgsForCall(0)
						Expect(service).To(Equal("my-service"))
						Expect(key).To(Equal("my-key"))
						Expect(spaceGUID).To(Equal("some-space-guid"))

						Expect(executeErr).NotTo(HaveOccurred())
						Expect(testUI.Out).To(Say("Creating service key my-key for service instance my-service as some-user..."))
						Expect(testUI.Out).To(Say("OK"))
					})
				})

				When("the service key creation fails", func() {
					BeforeEach(func() {
						fakeActor.CreateServiceKeyReturns(v2action.ServiceKey{}, nil, errors.New("explode"))
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError("explode"))
					})
				})

				When("the service key already exists", func() {
					BeforeEach(func() {
						fakeActor.CreateServiceKeyReturns(v2action.ServiceKey{}, nil, ccerror.ServiceKeyTakenError{})
					})

					It("displays OK and the warning", func() {
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Err).To(Say("Service key my-key already exists"))
					})

					It("does not return an error", func() {
						Expect(executeErr).NotTo(HaveOccurred())
					})
				})
			})

			When("passed JSON parameters", func() {
				var params map[string]interface{}

				BeforeEach(func() {
					params = map[string]interface{}{"foo": "bar"}
					cmd.ParametersAsJSON = params
				})

				It("behaves as usual, passing on the params", func() {
					Expect(testUI.Out).To(Say("Creating service key my-key for service instance my-service as some-user..."))

					Expect(fakeActor.CreateServiceKeyCallCount()).To(Equal(1))
					_, _, _, expectedParams := fakeActor.CreateServiceKeyArgsForCall(0)
					Expect(expectedParams).To(Equal(params))
				})
			})
		})
	})
})
