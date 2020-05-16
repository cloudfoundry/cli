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

var _ = Describe("router-groups Command", func() {
	var (
		cmd             RouterGroupsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
		args            []string
		binaryName      string
	)

	const tableHeaders = `name\s+type`

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		args = nil

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = RouterGroupsCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(args)
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

	Context("when the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
		})

		When("getting router groups succeeds", func() {
			var (
				routerGroups []v7action.RouterGroup
			)

			BeforeEach(func() {
				routerGroups = []v7action.RouterGroup{
					{Name: "rg1", Type: "type1"},
					{Name: "rg2", Type: "type2"},
					{Name: "rg3", Type: "type3"},
				}

				fakeActor.GetRouterGroupsReturns(
					routerGroups,
					nil,
				)
			})

			It("prints flavor text", func() {
				Expect(testUI.Out).To(Say("Getting router groups as banana..."))
			})

			It("prints routes in a table", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say(tableHeaders))
				Expect(testUI.Out).To(Say(`rg1\s+type1\s+`))
				Expect(testUI.Out).To(Say(`rg2\s+type2\s+`))
				Expect(testUI.Out).To(Say(`rg3\s+type3\s+`))
			})
		})

		When("getting router groups succeeds, but there are no groups", func() {
			BeforeEach(func() {
				fakeActor.GetRouterGroupsReturns(
					[]v7action.RouterGroup{},
					nil,
				)
			})

			It("prints flavor text", func() {
				Expect(testUI.Out).To(Say("Getting router groups as banana..."))
			})

			It("displays an empty message", func() {
				Expect(testUI.Out).To(Say("No router groups found."))
			})
		})

		When("getting router groups fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some-error")
				fakeActor.GetRouterGroupsReturns(nil, expectedErr)
			})

			It("prints flavor text", func() {
				Expect(testUI.Out).To(Say("Getting router groups as banana..."))
			})

			It("prints warnings and returns error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(testUI.Out).ToNot(Say(tableHeaders))
			})
		})
	})
})
