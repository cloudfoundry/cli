package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("service command", func() {
	const (
		serviceInstanceName = "fake-service-instance-name"
		serviceInstanceGUID = "fake-service-instance-guid"
		spaceGUID           = "fake-space-guid"
	)
	var (
		cmd             ServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
	)

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = ServiceCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{
			GUID: spaceGUID,
		})

		//fakeConfig.TargetedOrganizationReturns(configv3.Organization{
		//	GUID: "fake-org-guid",
		//	Name: "fake-org-name",
		//})
		//
		//fakeConfig.CurrentUserReturns(configv3.User{Name: "fake-username"}, nil)

		fakeActor.GetServiceInstanceByNameAndSpaceReturns(
			resources.ServiceInstance{
				GUID: serviceInstanceGUID,
				Name: serviceInstanceName,
			},
			v7action.Warnings{"warning one", "warning two"},
			nil,
		)

		setPositionalFlags(&cmd, serviceInstanceName)
		setFlag(&cmd, "--guid")
	})

	It("checks the user is logged in, and targeting an org and space", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		orgChecked, spaceChecked := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(orgChecked).To(BeTrue())
		Expect(spaceChecked).To(BeTrue())
	})

	It("looks up the service instance and prints the GUID and warnings", func() {
		Expect(executeErr).NotTo(HaveOccurred())

		Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
		actualName, actualSpaceGUID := fakeActor.GetServiceInstanceByNameAndSpaceArgsForCall(0)
		Expect(actualName).To(Equal(serviceInstanceName))
		Expect(actualSpaceGUID).To(Equal(spaceGUID))

		Expect(testUI.Out).To(Say(`^%s\n$`, serviceInstanceGUID))
		Expect(testUI.Err).To(SatisfyAll(
			Say("warning one"),
			Say("warning two"),
		))
	})

	When("there is a problem looking up the service instance", func() {
		BeforeEach(func() {
			fakeActor.GetServiceInstanceByNameAndSpaceReturns(
				resources.ServiceInstance{},
				v7action.Warnings{"warning one", "warning two"},
				errors.New("boom"),
			)
		})

		It("prints warnings and returns an error", func() {
			Expect(executeErr).To(MatchError("boom"))

			Expect(testUI.Out).NotTo(Say(`.`), "output not empty!")
			Expect(testUI.Err).To(SatisfyAll(
				Say("warning one"),
				Say("warning two"),
			))
		})
	})

	When("checking the target returns an error", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("explode"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("explode"))
		})
	})
})
