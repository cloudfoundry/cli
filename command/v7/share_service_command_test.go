package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("share-service Command", func() {
	var (
		cmd             ShareServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = ShareServiceCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("user not targeting space", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("space not targeted"))
		})

		It("checks the user is logged in, and targeting an org and space", func() {
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			orgChecked, spaceChecked := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(orgChecked).To(BeTrue())
			Expect(spaceChecked).To(BeTrue())
		})

		It("fails the command", func() {
			Expect(executeErr).To(Not(BeNil()))
			Expect(executeErr.Error()).To(ContainSubstring("space not targeted"))
		})
	})

})
