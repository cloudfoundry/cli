package v6_test

import (
	"errors"
	"regexp"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
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

var _ = Describe("v3-apply-manifest Command", func() {
	var (
		cmd             V3ApplyManifestCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeV3ApplyManifestActor
		fakeParser      *v6fakes.FakeManifestParser
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeV3ApplyManifestActor)
		fakeParser = new(v6fakes.FakeManifestParser)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = V3ApplyManifestCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			Parser:      fakeParser,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("displays the experimental warning", func() {
		Expect(testUI.Err).To(Say("This command is in EXPERIMENTAL stage and may change without notice"))
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("the user is logged in", func() {
		var (
			providedPath string
		)

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
				GUID: "some-space-guid",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)

			providedPath = "some-manifest-path"
			cmd.PathToManifest = flag.PathWithExistenceCheck(providedPath)
		})

		When("the parse is successful", func() {
			BeforeEach(func() {
				fakeActor.ApplyApplicationManifestReturns(
					v3action.Warnings{"some-manifest-warning"},
					nil,
				)
			})

			It("displays the success text", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Out).To(Say("Applying manifest %s in org some-org / space some-space as steve...", regexp.QuoteMeta(providedPath)))
				Expect(testUI.Err).To(Say("some-manifest-warning"))
				Expect(testUI.Out).To(Say("OK"))

				Expect(fakeParser.InterpolateAndParseCallCount()).To(Equal(1))
				path, _, _ := fakeParser.InterpolateAndParseArgsForCall(0)
				Expect(path).To(Equal(providedPath))

				Expect(fakeActor.ApplyApplicationManifestCallCount()).To(Equal(1))
				parserArg, spaceGUIDArg := fakeActor.ApplyApplicationManifestArgsForCall(0)
				Expect(parserArg).To(Equal(fakeParser))
				Expect(spaceGUIDArg).To(Equal("some-space-guid"))
			})
		})

		When("the parse errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("oooooh nooooos")
				fakeParser.InterpolateAndParseReturns(expectedErr)
			})

			It("returns back the parse error", func() {
				Expect(executeErr).To(MatchError(expectedErr))

				Expect(fakeActor.ApplyApplicationManifestCallCount()).To(Equal(0))
			})
		})
	})
})
