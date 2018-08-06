package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("rename buildpack command", func() {
	var (
		cmd             RenameBuildpackCommand
		fakeActor       *v2fakes.FakeRenameBuildpackActor
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		testUI          *ui.UI

		// bpName string
		// buildpack  v2action.Buildpack
		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeRenameBuildpackActor)
		// bpName = "some-new-bp-name"

		cmd = RenameBuildpackCommand{
			UI:          testUI,
			Actor:       fakeActor,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			// RequiredArgs: flag.RenameBuildpackArgs{},
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking the target fails", func() {
		var binaryName string

		BeforeEach(func() {
			binaryName = "faceman"
			fakeConfig.BinaryNameReturns(binaryName)
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "faceman"}))
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	Context("when checking the target succeeds", func() {
		var (
			oldName string
			newName string
		)

		BeforeEach(func() {
			oldName = "some-old-name"
			newName = "some-new-name"

			cmd.RequiredArgs = flag.RenameBuildpackArgs{
				OldBuildpackName: oldName,
				NewBuildpackName: newName,
			}
		})

		Context("when renaming to a unique buildpack name", func() {
			BeforeEach(func() {
				fakeActor.GetBuildpackByNameReturns(v2action.Buildpack{
					Name: oldName,
					GUID: "some-guid",
				}, v2action.Warnings{"warning1", "warning2"}, nil)
				fakeActor.UpdateBuildpackReturns(v2action.Buildpack{
					Name: newName,
					GUID: "some-guid",
				}, v2action.Warnings{"warning3", "warning4"}, nil)
			})

			It("successfully renames the buildpack and displays any warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeActor.GetBuildpackByNameCallCount()).To(Equal(1))
				Expect(fakeActor.GetBuildpackByNameArgsForCall(0)).To(Equal(oldName))

				Expect(fakeActor.UpdateBuildpackCallCount()).To(Equal(1))
				Expect(fakeActor.UpdateBuildpackArgsForCall(0)).To(Equal(v2action.Buildpack{
					Name: newName,
					GUID: "some-guid",
				}))

				Expect(testUI.Out).To(Say("Renaming buildpack %s to %s...", oldName, newName))
				Expect(testUI.Err).To(Say("warning1\\nwarning2"))
				Expect(testUI.Err).To(Say("warning3\\nwarning4"))
			})
		})

		// Context("when renaming to the same name", func() {
		// 	BeforeEach(func() {
		// 		fakeActor.GetBuildpackByNameReturns(v2action.Buildpack{
		// 			Name: oldName,
		// 			GUID: "some-guid",
		// 		}, v2action.Warnings{"warning1", "warning2"}, nil)
		// 		fakeActor.UpdateBuildpackReturns(v2action.Buildpack{
		// 			Name: newName,
		// 			GUID: "some-guid",
		// 		}, v2action.Warnings{"warning3", "warning4"}, nil)
		// 	})

		// 	It("successfully renames the buildpack", func() {
		// 	})
		// })

		Context("when the actor returns a multiple buildpacks found error", func() {
			BeforeEach(func() {
				fakeActor.GetBuildpackByNameReturns(v2action.Buildpack{}, v2action.Warnings{"warning1", "warning2"}, actionerror.MultipleBuildpacksFoundError{BuildpackName: "some-bp-name"})
			})

			It("returns an error and prints warnings", func() {
				Expect(executeErr).To(MatchError(translatableerror.MultipleBuildpacksFoundError{BuildpackName: "some-bp-name"}))
				Expect(testUI.Err).To(Say("warning1\\nwarning2"))
			})
		})

		Context("when the actor errors when looking up the buildpack", func() {
			BeforeEach(func() {
				fakeActor.GetBuildpackByNameReturns(v2action.Buildpack{}, v2action.Warnings{"warning1", "warning2"}, errors.New("buildpack error!"))
			})

			It("returns an error and prints warnings", func() {
				Expect(executeErr).To(MatchError("buildpack error!"))
				Expect(testUI.Err).To(Say("warning1\\nwarning2"))
			})
		})

		Context("when the actor errors because the target buildpack does not exist", func() {
			BeforeEach(func() {
				fakeActor.GetBuildpackByNameReturns(v2action.Buildpack{}, v2action.Warnings{"warning1", "warning2"}, actionerror.BuildpackNotFoundError{BuildpackName: oldName})
			})

			It("returns an error and prints warnings", func() {
				Expect(executeErr).To(MatchError(translatableerror.BuildpackNotFoundError{BuildpackName: oldName}))
				Expect(testUI.Err).To(Say("warning1\\nwarning2"))
			})
		})

		Context("when the actor errors because a buildpack with the desired name already exists", func() {
			BeforeEach(func() {
				fakeActor.UpdateBuildpackReturns(v2action.Buildpack{}, v2action.Warnings{"warning1", "warning2"}, actionerror.BuildpackNameTakenError(newName))
			})

			It("returns an error and prints warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.BuildpackNameTakenError(newName)))
				Expect(testUI.Err).To(Say("warning1\\nwarning2"))
			})
		})

		Context("when the actor errors when updating the buildpack", func() {
			BeforeEach(func() {
				fakeActor.UpdateBuildpackReturns(v2action.Buildpack{}, v2action.Warnings{"warning1", "warning2"}, errors.New("update error!"))
			})

			It("returns an error and prints warnings", func() {
				Expect(executeErr).To(MatchError("update error!"))
				Expect(testUI.Err).To(Say("warning1\\nwarning2"))
			})
		})
	})
})
