package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/shared/sharedfakes"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("copy-source Command", func() {
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
		sourceApp     resources.Application
		targetApp     resources.Application
		actorError    error
		targetPackage resources.Package
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		fakeAppStager = new(sharedfakes.FakeAppStager)

		binaryName = "faceman"
		userName = "banana"
		sourceApp = resources.Application{
			Name: "source-app-name",
			GUID: "source-app-guid",
		}
		targetApp = resources.Application{
			Name: "target-app-name",
			GUID: "target-app-guid",
		}
		targetPackage = resources.Package{
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
		fakeActor.GetCurrentUserReturns(configv3.User{Name: userName}, nil)

		fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})

		fakeActor.GetSpaceByNameAndOrganizationReturns(resources.Space{Name: "destination-space", GUID: "destination-space-guid"},
			v7action.Warnings{"get-space-by-name-warning"},
			nil,
		)
		fakeActor.GetOrganizationByNameReturns(resources.Organization{Name: "destination-org", GUID: "destination-org-guid"},
			v7action.Warnings{"get-org-by-name-warning"},
			nil,
		)

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
		Expect(fakeActor.GetCurrentUserCallCount()).To(Equal(1))
	})

	When("retrieving the current user fails", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("not-logged-in"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("not-logged-in"))
		})
	})

	When("the target organization is specified but the targeted space isn't", func() {
		BeforeEach(func() {
			cmd.Organization = "some-other-organization"
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.RequiredFlagsError{
				Arg1: "--organization, -o",
				Arg2: "--space, -s",
			}))
		})
	})

	When("the no restart and strategy flags are both provided", func() {
		BeforeEach(func() {
			cmd.NoRestart = true
			cmd.Strategy = flag.DeploymentStrategy{Name: constant.DeploymentStrategyRolling}
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
				Args: []string{
					"--no-restart", "--strategy",
				},
			}))
		})
	})

	When("the no restart and no wait flags are both provided", func() {
		BeforeEach(func() {
			cmd.NoRestart = true
			cmd.NoWait = true
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
				Args: []string{
					"--no-restart", "--no-wait",
				},
			}))
		})
	})

	When("a target org and space is provided", func() {
		BeforeEach(func() {
			cmd.Organization = "destination-org"
			cmd.Space = "destination-space"
		})

		It("retrieves the org by name and the space by name and organization", func() {
			Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
			org := fakeActor.GetOrganizationByNameArgsForCall(0)
			Expect(org).To(Equal(cmd.Organization))

			Expect(fakeActor.GetSpaceByNameAndOrganizationCallCount()).To(Equal(1))
			space, orgGUID := fakeActor.GetSpaceByNameAndOrganizationArgsForCall(0)
			Expect(space).To(Equal(cmd.Space))
			Expect(orgGUID).To(Equal("destination-org-guid"))
		})

		It("displays the warnings", func() {
			Expect(testUI.Err).To(Say("get-org-by-name-warning"))
		})

		When("retrieving the organization fails", func() {
			BeforeEach(func() {
				fakeActor.GetOrganizationByNameReturns(
					resources.Organization{},
					v7action.Warnings{},
					errors.New("get-org-by-name-err"),
				)
			})
			It("returns an error", func() {
				Expect(executeErr).To(MatchError("get-org-by-name-err"))
			})
		})

		When("retrieving the space fails", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceByNameAndOrganizationReturns(
					resources.Space{},
					v7action.Warnings{},
					errors.New("get-space-by-name-err"),
				)
			})
			It("returns an error", func() {
				Expect(executeErr).To(MatchError("get-space-by-name-err"))
			})
		})

		It("uses the provided org and space", func() {
			Expect(testUI.Out).To(Say(
				"Copying source from app %s to target app %s in org %s / space %s as %s...",
				sourceApp.Name,
				targetApp.Name,
				"destination-org",
				"destination-space",
				userName,
			))
		})
	})

	When("only a target space is provided", func() {
		BeforeEach(func() {
			cmd.Space = "destination-space"
		})

		It("retrieves the space by name and organization", func() {
			Expect(fakeActor.GetSpaceByNameAndOrganizationCallCount()).To(Equal(1))
			space, orgGUID := fakeActor.GetSpaceByNameAndOrganizationArgsForCall(0)
			Expect(space).To(Equal(cmd.Space))
			Expect(orgGUID).To(Equal(fakeConfig.TargetedOrganization().GUID))
		})

		It("displays the warnings", func() {
			Expect(testUI.Err).To(Say("get-space-by-name-warning"))
		})

		It("uses the provided org and space", func() {
			Expect(testUI.Out).To(Say(
				"Copying source from app %s to target app %s in org %s / space %s as %s...",
				sourceApp.Name,
				targetApp.Name,
				fakeConfig.TargetedOrganization().Name,
				"destination-space",
				userName,
			))
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
			fakeActor.GetApplicationByNameAndSpaceReturnsOnCall(0, resources.Application{}, v7action.Warnings{"get-source-app-warning"}, errors.New("get-source-app-error"))
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
			fakeActor.GetApplicationByNameAndSpaceReturnsOnCall(1, resources.Application{}, v7action.Warnings{"get-target-app-warning"}, errors.New("get-target-app-error"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("get-target-app-error"))
		})
	})

	It("copies the package", func() {
		Expect(fakeActor.CopyPackageCallCount()).To(Equal(1))
		srcApp, tgtApp := fakeActor.CopyPackageArgsForCall(0)
		Expect(srcApp).To(Equal(sourceApp))
		Expect(tgtApp).To(Equal(targetApp))

		Expect(testUI.Err).To(Say("copy-package-warning"))
	})

	When("copying the package fails", func() {
		BeforeEach(func() {
			actorError = errors.New("copy-package-error")
			fakeActor.CopyPackageReturns(resources.Package{}, v7action.Warnings{}, actorError)
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actorError))
		})
	})

	When("the strategy flag is set to rolling", func() {
		BeforeEach(func() {
			cmd.Strategy = flag.DeploymentStrategy{
				Name: constant.DeploymentStrategyRolling,
			}
		})

		It("stages and starts the app with the appropriate strategy", func() {
			Expect(fakeAppStager.StageAndStartCallCount()).To(Equal(1))
			returnedApp, spaceForApp, orgForApp, pkgGUID, opts := fakeAppStager.StageAndStartArgsForCall(0)
			Expect(returnedApp).To(Equal(targetApp))
			Expect(spaceForApp).To(Equal(configv3.Space{Name: "some-space", GUID: "some-space-guid"}))
			Expect(orgForApp).To(Equal(configv3.Organization{Name: "some-org"}))
			Expect(pkgGUID).To(Equal("target-package-guid"))
			Expect(opts.Strategy).To(Equal(constant.DeploymentStrategyRolling))
			Expect(opts.NoWait).To(Equal(false))
			Expect(opts.AppAction).To(Equal(constant.ApplicationRestarting))
		})
	})

	When("the no-wait flag is set", func() {
		BeforeEach(func() {
			cmd.NoWait = true
		})

		It("stages and starts the app with the appropriate strategy", func() {
			Expect(fakeAppStager.StageAndStartCallCount()).To(Equal(1))
			returnedApp, spaceForApp, orgForApp, pkgGUID, opts := fakeAppStager.StageAndStartArgsForCall(0)
			Expect(returnedApp).To(Equal(targetApp))
			Expect(spaceForApp).To(Equal(configv3.Space{Name: "some-space", GUID: "some-space-guid"}))
			Expect(orgForApp).To(Equal(configv3.Organization{Name: "some-org"}))
			Expect(pkgGUID).To(Equal("target-package-guid"))
			Expect(opts.Strategy).To(Equal(constant.DeploymentStrategyDefault))
			Expect(opts.NoWait).To(Equal(true))
			Expect(opts.AppAction).To(Equal(constant.ApplicationRestarting))
		})
	})

	It("stages and starts the target app", func() {
		Expect(fakeAppStager.StageAndStartCallCount()).To(Equal(1))
		returnedApp, spaceForApp, orgForApp, pkgGUID, opts := fakeAppStager.StageAndStartArgsForCall(0)
		Expect(returnedApp).To(Equal(targetApp))
		Expect(spaceForApp).To(Equal(configv3.Space{Name: "some-space", GUID: "some-space-guid"}))
		Expect(orgForApp).To(Equal(configv3.Organization{Name: "some-org"}))
		Expect(pkgGUID).To(Equal("target-package-guid"))
		Expect(opts.Strategy).To(Equal(constant.DeploymentStrategyDefault))
		Expect(opts.NoWait).To(Equal(false))
		Expect(opts.AppAction).To(Equal(constant.ApplicationRestarting))
	})

	When("staging and starting the app fails", func() {
		BeforeEach(func() {
			fakeAppStager.StageAndStartReturns(errors.New("stage-and-start-error"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("stage-and-start-error"))
		})
	})

	When("the no-restart flag is set", func() {
		BeforeEach(func() {
			cmd.NoRestart = true
		})
		It("succeeds but does not restart the app", func() {
			Expect(executeErr).To(Not(HaveOccurred()))
			Expect(fakeAppStager.StageAndStartCallCount()).To(Equal(0))
		})
	})

	It("succeeds", func() {
		Expect(executeErr).To(Not(HaveOccurred()))
	})
})
