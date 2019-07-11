package v6_test

import (
	"encoding/json"
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("update-user-provided-service Command", func() {
	const (
		fakeServiceInstanceName = "fake-service-instance-name"
	)

	var (
		cmd             *UpdateUserProvidedServiceCommand
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeUpdateUserProvidedServiceActor
		input           *Buffer
		testUI          *ui.UI
		executeErr      error
		extraArgs       []string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(v6fakes.FakeUpdateUserProvidedServiceActor)
		fakeSharedActor = new(commandfakes.FakeSharedActor)

		cmd = &UpdateUserProvidedServiceCommand{
			UI:           testUI,
			Config:       fakeConfig,
			SharedActor:  fakeSharedActor,
			Actor:        fakeActor,
			RequiredArgs: flag.ServiceInstance{ServiceInstance: fakeServiceInstanceName},
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(extraArgs)
	})

	It("checks the user is logged in, and targeting an org and space", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		orgChecked, spaceChecked := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(orgChecked).To(BeTrue())
		Expect(spaceChecked).To(BeTrue())
	})

	When("checking the target returns an error", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("explode"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("explode"))
		})
	})

	When("the user is logged in, and is targeting an org and space", func() {
		const (
			fakeOrgName   = "fake-org-name"
			fakeSpaceName = "fake-space-name"
			fakeSpaceGUID = "fake-space-guid"
			fakeUserName  = "fake-user-name"
		)

		BeforeEach(func() {
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: fakeSpaceName,
				GUID: fakeSpaceGUID,
			})

			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: fakeOrgName,
			})

			fakeConfig.CurrentUserReturns(configv3.User{Name: fakeUserName}, nil) //TODO: test errors
		})

		It("looks up the service instance GUID", func() {
			Expect(fakeActor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))
			name, spaceGUID := fakeActor.GetServiceInstanceByNameAndSpaceArgsForCall(0)
			Expect(name).To(Equal(fakeServiceInstanceName))
			Expect(spaceGUID).To(Equal(fakeSpaceGUID))
		})

		When("looking up the service instance GUID returns warnings", func() {
			BeforeEach(func() {
				fakeActor.GetServiceInstanceByNameAndSpaceReturns(
					v2action.ServiceInstance{},
					v2action.Warnings{"something obstreperous"},
					nil,
				)
			})

			It("reports the warning", func() {
				Expect(testUI.Err).To(Say("something obstreperous"))
			})
		})

		When("looking up the service instance GUID fails", func() {
			BeforeEach(func() {
				fakeActor.GetServiceInstanceByNameAndSpaceReturns(
					v2action.ServiceInstance{},
					v2action.Warnings{"something obstreperous"},
					errors.New("something awful"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(testUI.Err).To(Say("something obstreperous"))
				Expect(executeErr).To(MatchError("something awful"))
			})
		})

		When("the service instance is not user-provided", func() {
			It("fails with an error", func() {
				Expect(executeErr).To(MatchError(fmt.Sprintf("The service instance '%s' is not user-provided", fakeServiceInstanceName)))
			})
		})

		When("looking up the service instance GUID succeeds, and it is user-provided", func() {
			const (
				fakeServiceInstanceGUID = "fake-service-instance-guid"
			)

			BeforeEach(func() {
				fakeActor.GetServiceInstanceByNameAndSpaceReturns(
					v2action.ServiceInstance{
						GUID: fakeServiceInstanceGUID,
						Type: constant.UserProvidedService,
					},
					nil,
					nil,
				)
			})

			When("no flags are provided", func() {
				It("succeeds", func() {
					Expect(executeErr).NotTo(HaveOccurred())
				})

				It("says that no flags were provided", func() {
					Expect(testUI.Out).To(Say("No flags specified. No changes were made"))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			Context("updating log URL", func() {
				BeforeEach(func() {
					cmd.SyslogDrainURL.IsSet = true
					cmd.SyslogDrainURL.Value = "fake-syslog-drain-url"
				})

				It("succeeds", func() {
					Expect(executeErr).NotTo(HaveOccurred())
				})

				It("displays messages to the user", func() {
					expectOKMessage(testUI, fakeServiceInstanceName, fakeOrgName, fakeSpaceName, fakeUserName)
				})

				It("updates the syslog URL", func() {
					Expect(fakeActor.UpdateUserProvidedServiceInstanceCallCount()).To(Equal(1))
					guid, instanceChanges := fakeActor.UpdateUserProvidedServiceInstanceArgsForCall(0)
					Expect(guid).To(Equal(fakeServiceInstanceGUID))
					Expect(instanceChanges).To(MatchAllFields(Fields{
						"SyslogDrainURL":  PointTo(Equal("fake-syslog-drain-url")),
						"RouteServiceURL": BeNil(),
						"Tags":            BeNil(),
						"Credentials":     BeNil(),
					}))
				})
			})

			Context("updating credentials", func() {
				BeforeEach(func() {
					cmd.Credentials.IsSet = true
					cmd.Credentials.UserPromptCredentials = []string{"pass phrase", "cred"}

					input.Write([]byte("very secret passphrase\nsecret cred\n"))
				})

				It("prompts the user for credentials", func() {
					Expect(testUI.Out).To(Say("pass phrase: "))
					Expect(testUI.Out).To(Say("cred: "))
				})

				It("does not echo the credentials", func() {
					Expect(testUI.Out).NotTo(Say("secret"))
					Expect(testUI.Err).NotTo(Say("secret"))
				})

				It("succeeds", func() {
					Expect(executeErr).NotTo(HaveOccurred())
				})

				It("displays messages to the user", func() {
					expectOKMessage(testUI, fakeServiceInstanceName, fakeOrgName, fakeSpaceName, fakeUserName)
				})

				It("updates the credentials", func() {
					Expect(fakeActor.UpdateUserProvidedServiceInstanceCallCount()).To(Equal(1))
					guid, instanceChanges := fakeActor.UpdateUserProvidedServiceInstanceArgsForCall(0)
					Expect(guid).To(Equal(fakeServiceInstanceGUID))
					Expect(instanceChanges).To(MatchAllFields(Fields{
						"Credentials": Equal(map[string]interface{}{
							"pass phrase": "very secret passphrase",
							"cred":        "secret cred",
						}),
						"RouteServiceURL": BeNil(),
						"Tags":            BeNil(),
						"SyslogDrainURL":  BeNil(),
					}))
				})
			})

			Context("updating routes URL", func() {
				BeforeEach(func() {
					cmd.RouteServiceURL.IsSet = true
					cmd.RouteServiceURL.Value = "fake-route-url"
				})

				It("succeeds", func() {
					Expect(executeErr).NotTo(HaveOccurred())
				})

				It("displays messages to the user", func() {
					expectOKMessage(testUI, fakeServiceInstanceName, fakeOrgName, fakeSpaceName, fakeUserName)
				})

				It("updates the routes URL", func() {
					Expect(fakeActor.UpdateUserProvidedServiceInstanceCallCount()).To(Equal(1))
					guid, instanceChanges := fakeActor.UpdateUserProvidedServiceInstanceArgsForCall(0)
					Expect(guid).To(Equal(fakeServiceInstanceGUID))
					Expect(instanceChanges).To(MatchAllFields(Fields{
						"RouteServiceURL": PointTo(Equal("fake-route-url")),
						"Tags":            BeNil(),
						"SyslogDrainURL":  BeNil(),
						"Credentials":     BeNil(),
					}))
				})
			})

			Context("updating tags", func() {
				BeforeEach(func() {
					cmd.Tags.IsSet = true
					cmd.Tags.Value = []string{"foo", "bar"}
				})

				It("succeeds", func() {
					Expect(executeErr).NotTo(HaveOccurred())
				})

				It("displays messages to the user", func() {
					expectOKMessage(testUI, fakeServiceInstanceName, fakeOrgName, fakeSpaceName, fakeUserName)
				})

				It("updates the tags", func() {
					Expect(fakeActor.UpdateUserProvidedServiceInstanceCallCount()).To(Equal(1))
					guid, instanceChanges := fakeActor.UpdateUserProvidedServiceInstanceArgsForCall(0)
					Expect(guid).To(Equal(fakeServiceInstanceGUID))
					Expect(instanceChanges).To(MatchAllFields(Fields{
						"Tags":            PointTo(ConsistOf("foo", "bar")),
						"SyslogDrainURL":  BeNil(),
						"RouteServiceURL": BeNil(),
						"Credentials":     BeNil(),
					}))
				})
			})

			When("unsetting values", func() {
				BeforeEach(func() {
					cmd.RouteServiceURL.IsSet = true
					cmd.Credentials.IsSet = true
					cmd.SyslogDrainURL.IsSet = true
					cmd.Tags.IsSet = true
				})

				It("sends empty values", func() {
					Expect(fakeActor.UpdateUserProvidedServiceInstanceCallCount()).To(Equal(1))
					guid, instanceChanges := fakeActor.UpdateUserProvidedServiceInstanceArgsForCall(0)
					Expect(guid).To(Equal(fakeServiceInstanceGUID))
					bytes, err := json.Marshal(instanceChanges)
					Expect(err).NotTo(HaveOccurred())
					Expect(bytes).To(MatchJSON(`
          {
					  "tags": [],
						"syslog_drain_url": "",
						"route_service_url": "",
						"credentials": {}
					}`))
				})
			})

			When("the action returns warnings", func() {
				BeforeEach(func() {
					cmd.Tags.IsSet = true
					cmd.Tags.Value = []string{"foo", "bar"}

					fakeActor.UpdateUserProvidedServiceInstanceReturns(v2action.Warnings{"some", "warnings"}, nil)
				})

				It("reports the warnings to the user", func() {
					Expect(testUI.Err).To(Say("some"))
					Expect(testUI.Err).To(Say("warnings"))
				})
			})

			When("the action fails", func() {
				BeforeEach(func() {
					cmd.Tags.IsSet = true
					cmd.Tags.Value = []string{"foo", "bar"}

					fakeActor.UpdateUserProvidedServiceInstanceReturns(
						v2action.Warnings{"some", "warnings"},
						errors.New("utterly awful happenings"),
					)
				})

				It("reports the failure and warnings", func() {
					Expect(testUI.Err).To(Say("some"))
					Expect(testUI.Err).To(Say("warnings"))
					Expect(executeErr).To(MatchError("utterly awful happenings"))
				})
			})
		})
	})
})

// var _ = Describe("update-service Command", func() {
//
// 	const (
// 		serviceInstanceName = "my-service"
// 		spaceGUID           = "space-guid"
// 		instanceGUID        = "instance-guid"
// 		planGUID            = "plan-guid"
// 	)
//
// 	var (
// 		cmd             UpdateServiceCommand
// 		fakeActor       *v6fakes.FakeUpdateServiceActor
// 		fakeSharedActor *commandfakes.FakeSharedActor
// 		fakeConfig      *commandfakes.FakeConfig
// 		testUI          *ui.UI
// 		input           *Buffer
// 		executeErr      error
// 		extraArgs       []string
//
// 		space = configv3.Space{Name: "space-a", GUID: spaceGUID}
// 	)
//
// 	BeforeEach(func() {
// 		input = NewBuffer()
// 		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
// 		fakeActor = new(v6fakes.FakeUpdateServiceActor)
// 		fakeSharedActor = new(commandfakes.FakeSharedActor)
// 		fakeConfig = new(commandfakes.FakeConfig)
//
// 		fakeConfig.TargetedSpaceReturns(space)
//
// 		extraArgs = []string{}
//
// 		cmd = UpdateServiceCommand{
// 			UI:           testUI,
// 			Actor:        fakeActor,
// 			SharedActor:  fakeSharedActor,
// 			Config:       fakeConfig,
// 			RequiredArgs: flag.ServiceInstance{ServiceInstance: serviceInstanceName},
// 		}
// 	})
//
// 	JustBeforeEach(func() {
// 		executeErr = cmd.Execute(extraArgs)
// 	})
//
// 	When("not upgrading", func() {
// 		It("returns UnrefactoredCommandError", func() {
// 			// delegates non-upgrades to legacy code
// 			Expect(executeErr).To(MatchError(translatableerror.UnrefactoredCommandError{}))
// 		})
// 	})
//
// 	When("combining upgrade with other flags", func() {
// 		BeforeEach(func() {
// 			cmd.Upgrade = true
// 		})
//
// 		When("tags provided", func() {
// 			BeforeEach(func() {
// 				cmd.Tags = "tags"
// 			})
//
// 			It("returns UpgradeArgumentCombinationError", func() {
// 				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
// 					Args: []string{"--upgrade", "-t", "-c", "-p"},
// 				}))
// 			})
// 		})
//
// 		When("parameters provided", func() {
// 			BeforeEach(func() {
// 				cmd.ParametersAsJSON = "{\"some\": \"stuff\"}"
// 			})
//
// 			It("returns UpgradeArgumentCombinationError", func() {
// 				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
// 					Args: []string{"--upgrade", "-t", "-c", "-p"},
// 				}))
// 			})
// 		})
//
// 		When("plan provided", func() {
// 			BeforeEach(func() {
// 				cmd.Plan = "new-plan"
// 			})
//
// 			It("returns UpgradeArgumentCombinationError", func() {
// 				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
// 					Args: []string{"--upgrade", "-t", "-c", "-p"},
// 				}))
// 			})
// 		})
// 	})
//
// 	When("upgrading", func() {
// 		BeforeEach(func() {
// 			cmd.Upgrade = true
// 		})
//
// 		When("the version of CC API is less than minimum version supporting maintenance_info updates", func() {
// 			BeforeEach(func() {
// 				fakeActor.CloudControllerAPIVersionReturns(ccversion.MinSupportedV2ClientVersion)
// 			})
//
// 			It("should warn the user that the version of CAPI is too low and exit with an error", func() {
// 				Expect(executeErr).To(MatchError(translatableerror.MinimumCFAPIVersionNotMetError{
// 					Command:        "Option '--upgrade'",
// 					CurrentVersion: ccversion.MinSupportedV2ClientVersion,
// 					MinimumVersion: ccversion.MinVersionUpdateServiceInstanceMaintenanceInfoV2,
// 				}))
// 			})
// 		})
//
// 		When("the version of CC API supports maintenance_info updates", func() {
// 			BeforeEach(func() {
// 				fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionUpdateServiceInstanceMaintenanceInfoV2)
// 			})
//
// 			It("checks the user is logged in, and targeting an org and space", func() {
// 				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
// 				orgChecked, spaceChecked := fakeSharedActor.CheckTargetArgsForCall(0)
// 				Expect(orgChecked).To(BeTrue())
// 				Expect(spaceChecked).To(BeTrue())
// 			})
//
// 			When("checking the target succeeds", func() {
// 				When("getting the service instance succeeds", func() {
// 					BeforeEach(func() {
// 						fakeActor.GetServiceInstanceByNameAndSpaceReturns(
// 							v2action.ServiceInstance{GUID: instanceGUID, ServicePlanGUID: planGUID},
// 							v2action.Warnings{"warning"},
// 							nil)
// 					})
//
// 					It("displays any warnings", func() {
// 						Expect(testUI.Err).To(Say("warning"))
// 					})
//
// 					It("mentions that the command is experimental", func() {
// 						Expect(testUI.Out).To(Say("This command is in EXPERIMENTAL stage and may change without notice\\."))
// 					})
//
// 					It("prompts the user about the upgrade", func() {
// 						Expect(testUI.Out).To(Say("You are about to update %s\\.", serviceInstanceName))
// 						Expect(testUI.Out).To(Say("Warning: This operation may be long running and will block further operations on the service until complete\\."))
// 						Expect(testUI.Out).To(Say("Really update service %s\\? \\[yN\\]:", serviceInstanceName))
// 					})
//
// 					When("user refuses to proceed with the upgrade", func() {
// 						BeforeEach(func() {
// 							input.Write([]byte("n\n"))
// 						})
//
// 						It("does not send an upgrade request", func() {
// 							Expect(fakeActor.UpgradeServiceInstanceCallCount()).To(Equal(0))
// 						})
//
// 						It("cancels the update", func() {
// 							Expect(executeErr).NotTo(HaveOccurred())
// 							Expect(testUI.Out).To(Say("Update cancelled"))
// 						})
// 					})
//
// 					When("user goes ahead with the upgrade", func() {
// 						BeforeEach(func() {
// 							input.Write([]byte("y\n"))
// 						})
//
// 						It("sends an upgrade request", func() {
// 							Expect(fakeActor.UpgradeServiceInstanceCallCount()).To(Equal(1), "upgrade should be requested")
//
// 							serviceInstanceGUID, servicePlanGUID := fakeActor.UpgradeServiceInstanceArgsForCall(0)
// 							Expect(serviceInstanceGUID).To(Equal(instanceGUID))
// 							Expect(servicePlanGUID).To(Equal(planGUID))
// 						})
//
// 						When("the update request succeeds", func() {
// 							It("says that the update was successful", func() {
// 								Expect(executeErr).NotTo(HaveOccurred())
// 								Expect(testUI.Out).To(Say("OK"))
// 							})
// 						})
//
// 						When("the update request fails", func() {
// 							BeforeEach(func() {
// 								fakeActor.UpgradeServiceInstanceReturns(
// 									v2action.Warnings{},
// 									fmt.Errorf("bad things happened"),
// 								)
// 							})
//
// 							It("says that the update has failed", func() {
// 								Expect(executeErr).To(MatchError("bad things happened"))
// 							})
// 						})
//
// 						When("there are warnings", func() {
// 							BeforeEach(func() {
// 								fakeActor.UpgradeServiceInstanceReturns(
// 									v2action.Warnings{"fake upgrade warning 1", "fake upgrade warning 2"},
// 									nil,
// 								)
// 							})
//
// 							It("outputs the warnings", func() {
// 								Expect(testUI.Err).To(Say("fake upgrade warning 1"))
// 								Expect(testUI.Err).To(Say("fake upgrade warning 2"))
// 							})
//
// 							It("can still output OK", func() {
// 								Expect(testUI.Out).To(Say("OK"))
// 							})
// 						})
// 					})
//
// 					When("user presses return", func() {
// 						BeforeEach(func() {
// 							input.Write([]byte("\n"))
// 						})
//
// 						It("cancels the update", func() {
// 							Expect(testUI.Out).To(Say("Update cancelled"))
// 							Expect(executeErr).NotTo(HaveOccurred())
// 						})
// 					})
//
// 					When("user does not answer", func() {
// 						It("fails", func() {
// 							Expect(executeErr).To(MatchError("EOF"))
// 						})
// 					})
// 				})
//
// 				When("getting the service instance fails", func() {
// 					BeforeEach(func() {
// 						fakeActor.GetServiceInstanceByNameAndSpaceReturns(v2action.ServiceInstance{}, v2action.Warnings{"warning"}, errors.New("explode"))
// 					})
//
// 					It("propagates the error", func() {
// 						Expect(executeErr).To(MatchError("explode"))
// 					})
//
// 					It("displays any warnings", func() {
// 						Expect(testUI.Err).To(Say("warning"))
// 					})
// 				})
// 			})
//
// 			When("too many arguments are provided", func() {
// 				BeforeEach(func() {
// 					extraArgs = []string{"extra"}
// 				})
//
// 				It("returns a TooManyArgumentsError", func() {
// 					Expect(executeErr).To(MatchError(translatableerror.TooManyArgumentsError{
// 						ExtraArgument: "extra",
// 					}))
// 				})
// 			})
//
// 			When("checking the target returns an error", func() {
// 				BeforeEach(func() {
// 					fakeSharedActor.CheckTargetReturns(errors.New("explode"))
// 				})
//
// 				It("returns an error", func() {
// 					Expect(executeErr).To(MatchError("explode"))
// 				})
// 			})
// 		})
// 	})
// })
//
func expectOKMessage(testUI *ui.UI, serviceName, orgName, spaceName, userName string) {
	Expect(testUI.Out).To(Say("Updating user provided service %s in org %s / space %s as %s...", serviceName, orgName, spaceName, userName))
	Expect(testUI.Out).To(Say("OK"))
	Expect(testUI.Out).To(Say("TIP: Use 'cf restage' for any bound apps to ensure your env variable changes take effect"))
}
