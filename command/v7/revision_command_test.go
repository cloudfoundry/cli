package v7_test

import (
	"code.cloudfoundry.org/cli/v7/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/v7/command/v7"
	"code.cloudfoundry.org/cli/v7/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v7/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("revision Command", func() {
	var (
		cmd             v7.RevisionCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		appName         string

		out *Buffer
	)

	BeforeEach(func() {
		out = NewBuffer()
		testUI = ui.NewTestUI(nil, out, NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = v7.RevisionCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}
		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		appName = "some-app"

		cmd.RequiredArgs.AppName = appName
	})

	JustBeforeEach(func() {
		Expect(cmd.Execute(nil)).To(Succeed())
	})

	It("displays the experimental warning", func() {
		Expect(testUI.Err).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
	})
})
