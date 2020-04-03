package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/shared/sharedfakes"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-app Command", func() {
	var (
		cmd             v7.CopySourceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		fakeAppStager   *sharedfakes.FakeAppStager

		binaryName    string
		userName      string
		executeErr    error
		sourceApp     v7action.Application
		targetApp     v7action.Application
		actorError    error
		targetPackage v7action.Package
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		fakeAppStager = new(sharedfakes.FakeAppStager)

		binaryName = "faceman"
		userName = "banana"
		sourceApp = v7action.Application{
			Name: "source-app-name",
			GUID: "source-app-guid",
		}
		targetApp = v7action.Application{
			Name: "target-app-name",
			GUID: "target-app-guid",
		}
		targetPackage = v7action.Package{
			GUID: "target-package-guid",
		}

		cmd = v7.CopySourceCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.CopySourceArgs{SourceAppName: sourceApp.Name, TargetAppName: targetApp.Name},
			Stager:       fakeAppStager,
		}

		fakeConfig.BinaryNameReturns(binaryName)
		fakeSharedActor.CheckTargetReturns(nil)
		fakeConfig.CurrentUserReturns(configv3.User{Name: userName}, nil)
		fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})

		fakeActor.GetApplicationByNameAndSpaceReturnsOnCall(0, sourceApp, v7action.Warnings{"get-source-app-warning"}, nil)
		fakeActor.GetApplicationByNameAndSpaceReturnsOnCall(1, targetApp, v7action.Warnings{"get-target-app-warning"}, nil)
		fakeActor.CopyPackageReturns(targetPackage, v7action.Warnings{"copy-package-warning"}, nil)

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
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	It("retrieves the current user", func() {
		Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
	})

	When("retrieving the current user fails", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("not-logged-in"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("not-logged-in"))
		})
	})

	It("displays a message about copying the source", func() {
		Expect(testUI.Out).To(Say(
			"Copying source from app %s to target app %s in org %s / space %s as %s...",
			sourceApp.Name,
			targetApp.Name,
			fakeConfig.TargetedOrganization().Name,
			fakeConfig.TargetedSpace().Name,
			userName,
		))
	})

	It("retrieves the source app", func() {
		Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(2))
		givenAppName, givenSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
		Expect(givenAppName).To(Equal(sourceApp.Name))
		Expect(givenSpaceGUID).To(Equal("some-space-guid"))

		Expect(testUI.Err).To(Say("get-source-app-warning"))
	})

	When("retrieving the source app fails", func() {
		BeforeEach(func() {
			fakeActor.GetApplicationByNameAndSpaceReturnsOnCall(0, v7action.Application{}, v7action.Warnings{"get-source-app-warning"}, errors.New("get-source-app-error"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("get-source-app-error"))
		})
	})

	It("retrieves the target app", func() {
		Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(2))
		givenAppName, givenSpaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(1)
		Expect(givenAppName).To(Equal(targetApp.Name))
		Expect(givenSpaceGUID).To(Equal("some-space-guid"))

		Expect(testUI.Err).To(Say("get-target-app-warning"))
	})

	When("retrieving the target app fails", func() {
		BeforeEach(func() {
			fakeActor.GetApplicationByNameAndSpaceReturnsOnCall(1, v7action.Application{}, v7action.Warnings{"get-target-app-warning"}, errors.New("get-target-app-error"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("get-target-app-error"))
		})
	})

	It("copies the package", func() {
		Expect(fakeActor.CopyPackageCallCount()).To(Equal(1))
		srcAppGUID, tgtAppGUID := fakeActor.CopyPackageArgsForCall(0)
		Expect(srcAppGUID).To(Equal(sourceApp.GUID))
		Expect(tgtAppGUID).To(Equal(targetApp.GUID))

		Expect(testUI.Err).To(Say("copy-package-warning"))
	})

	When("copying the package fails", func() {
		BeforeEach(func() {
			actorError = errors.New("copy-package-error")
			fakeActor.CopyPackageReturns(v7action.Package{}, v7action.Warnings{}, actorError)
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actorError))
		})
	})

	It("stages and starts the target app", func() {
		Expect(fakeAppStager.StageAndStartCallCount()).To(Equal(1))
		returnedApp, pkgGUID, strategy, noWait := fakeAppStager.StageAndStartArgsForCall(0)
		Expect(returnedApp).To(Equal(targetApp))
		Expect(pkgGUID).To(Equal("target-package-guid"))
		Expect(strategy).To(Equal(constant.DeploymentStrategyDefault))
		Expect(noWait).To(Equal(false))
	})

	When("staging and starting the app fails", func() {
		BeforeEach(func() {
			fakeAppStager.StageAndStartReturns(errors.New("stage-and-start-error"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("stage-and-start-error"))
		})
	})

	It("succeeds", func() {
		Expect(executeErr).To(Not(HaveOccurred()))
	})
})
