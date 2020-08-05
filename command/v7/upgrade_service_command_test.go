package v7_test

import (
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("upgrade-service command", func() {
	const (
		serviceInstanceName = "fake-service-instance-name"
		spaceName           = "fake-space-name"
		spaceGUID           = "fake-space-guid"
		orgName             = "fake-org-name"
		username            = "fake-username"
	)

	var (
		input           *Buffer
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		cmd             UpgradeServiceCommand
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = UpgradeServiceCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		setPositionalFlags(&cmd, serviceInstanceName)

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: orgName})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: spaceName,
			GUID: spaceGUID,
		})
		fakeConfig.CurrentUserReturns(configv3.User{Name: username}, nil)
	})

	Describe("Execute", func() {
		It("fails with a not implemented", func() {
			executeErr := cmd.Execute(nil)

			By("checking the user is logged in, and targeting an org and space", func() {
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				orgChecked, spaceChecked := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(orgChecked).To(BeTrue())
				Expect(spaceChecked).To(BeTrue())
			})

			By("outputting not implemented", func() {
				Expect(executeErr).To(MatchError("WIP: Not yet implemented"))
			})
		})
	})

	When("checking the target returns an error", func() {
		It("returns the error", func() {
			fakeSharedActor.CheckTargetReturns(errors.New("explode"))
			executeErr := cmd.Execute(nil)
			Expect(executeErr).To(MatchError("explode"))
		})
	})
})
