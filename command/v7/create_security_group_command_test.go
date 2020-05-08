package v7_test

import (
	"encoding/json"
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-security-group Command", func() {
	var (
		cmd             v7.CreateSecurityGroupCommand
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

		cmd = v7.CreateSecurityGroupCommand{
			BaseCommand: command.BaseCommand{
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
		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
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

	When("the provided JSON is invalid", func() {
		BeforeEach(func() {
			fakeActor.CreateSecurityGroupReturns(
				v7action.Warnings{"create-security-group-warning"},
				&json.SyntaxError{})
		})

		It("returns a custom error and provides an example", func() {
			Expect(executeErr).To(HaveOccurred())
			Expect(executeErr).To(Equal(actionerror.SecurityGroupJsonSyntaxError{Path: "some-path"}))
			Expect(testUI.Err).To(Say("create-security-group-warning"))
		})
	})

	When("the provided JSON does not contain the required fields", func() {
		BeforeEach(func() {
			fakeActor.CreateSecurityGroupReturns(
				v7action.Warnings{"create-security-group-warning"},
				&json.UnmarshalTypeError{})
		})

		It("returns a custom error and provides an example", func() {
			Expect(executeErr).To(HaveOccurred())
			Expect(executeErr).To(Equal(actionerror.SecurityGroupJsonSyntaxError{Path: "some-path"}))
			Expect(testUI.Err).To(Say("create-security-group-warning"))
		})
	})

	When("the security group already exists", func() {
		BeforeEach(func() {
			fakeActor.CreateSecurityGroupReturns(
				v7action.Warnings{},
				ccerror.SecurityGroupAlreadyExists{Message: fmt.Sprintf("Security group with name '%s' already exists", securityGroupName)})
		})

		It("returns a custom error and provides an example", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(testUI.Err).To(Say("Security group with name '%s' already exists", securityGroupName))
			Expect(testUI.Out).To(Say("OK"))
		})
	})

	When("it can't create the security group", func() {
		BeforeEach(func() {
			fakeActor.CreateSecurityGroupReturns(
				v7action.Warnings{"create-sec-grp-warning"},
				errors.New("some-error"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(HaveOccurred())

			Expect(testUI.Err).To(Say("create-sec-grp-warning"))
		})
	})
})
