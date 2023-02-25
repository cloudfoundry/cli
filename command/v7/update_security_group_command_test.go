package v7_test

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("update-security-group Command", func() {
	var (
		cmd             v7.UpdateSecurityGroupCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor

		binaryName            string
		executeErr            error
		securityGroupName     string
		securityGroupFilePath flag.PathWithExistenceCheck
	)

	BeforeEach(func() {
		securityGroupName = "some-name"
		securityGroupFilePath = "some-path"
		binaryName = "faceman"
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())

		cmd = v7.UpdateSecurityGroupCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.SecurityGroupArgs{
				SecurityGroup:   securityGroupName,
				PathToJSONRules: securityGroupFilePath,
			},
		}
		fakeConfig.HasTargetedOrganizationReturns(false)
		fakeConfig.HasTargetedSpaceReturns(false)
		fakeActor.GetCurrentUserReturns(configv3.User{Name: "steve"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
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

	When("updating the security group", func() {
		BeforeEach(func() {
			fakeActor.UpdateSecurityGroupReturns(
				v7action.Warnings{"update-security-group-warning"},
				nil)
		})

		It("succeeds", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(testUI.Err).To(Say("update-security-group-warning"))
			Expect(fakeActor.UpdateSecurityGroupCallCount()).To(Equal(1))
		})
	})

	When("the provided JSON is invalid", func() {
		BeforeEach(func() {
			fakeActor.UpdateSecurityGroupReturns(
				v7action.Warnings{"update-security-group-warning"},
				&json.SyntaxError{})
		})

		It("returns a custom error and provides an example", func() {
			Expect(executeErr).To(HaveOccurred())
			Expect(executeErr).To(Equal(actionerror.SecurityGroupJsonSyntaxError{Path: "some-path"}))
			Expect(testUI.Err).To(Say("update-security-group-warning"))
		})
	})

	When("something goes wrong updating the security group", func() {
		BeforeEach(func() {
			fakeActor.UpdateSecurityGroupReturns(
				v7action.Warnings{"update-security-group-warning"},
				actionerror.SecurityGroupNotFoundError{},
			)
		})

		It("returns a helpful error", func() {
			Expect(executeErr).To(HaveOccurred())
			Expect(executeErr).To(Equal(actionerror.SecurityGroupNotFoundError{}))
			Expect(testUI.Err).To(Say("update-security-group-warning"))
		})
	})
})
