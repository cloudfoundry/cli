package v7_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	"code.cloudfoundry.org/cli/cf/errors"

	. "code.cloudfoundry.org/cli/command/v7"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("set-org-role Command", func() {
	var (
		cmd             SetOrgQuotaCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeSetOrgQuotaActor
		binaryName      string
		executeErr      error
		input           *Buffer
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeSetOrgQuotaActor)

		cmd = SetOrgQuotaCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	BeforeEach(func() {
		fakeConfig.CurrentUserReturns(configv3.User{Name: "current-user"}, nil)

		fakeActor.GetOrganizationByNameReturns(
			v7action.Organization{GUID: "some-org-guid", Name: "some-org-name"},
			v7action.Warnings{"get-org-warning"},
			nil,
		)

		fakeActor.GetUserReturns(
			v7action.User{GUID: "target-user-guid", Username: "target-user"},
			nil,
		)
	})

	When("creating the relationship succeeds", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.SetOrgQuotaArgs{ }
			cmd.RequiredArgs.Organization = "some-quota-name"
			cmd.RequiredArgs.OrganizationQuota = "target-user-name"

			fakeActor.CreateOrgRoleReturns(
				v7action.Warnings{"create-role-warning"},
				errors.New("create-role-error"),
			)
		})

		It("displays warnings and returns without error", func() {
			Expect(testUI.Err).To(Say("create-role-warning"))
			Expect(executeErr).To(MatchError("create-role-error"))
		})
	})
})

