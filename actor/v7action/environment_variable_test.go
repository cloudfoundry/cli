package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Environment Variable Actions", func() {
	Describe("GetEnvironmentVariableGroup", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
			fetchedEnvVariables       EnvironmentVariableGroup
			executeErr                error
			warnings                  Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil)
		})

		JustBeforeEach(func() {
			fetchedEnvVariables, warnings, executeErr = actor.GetEnvironmentVariableGroup(constant.StagingEnvironmentVariableGroup)
		})

		When("getting the environment variable group fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetEnvironmentVariableGroupReturns(
					nil,
					ccv3.Warnings{"get-env-var-group-warning"},
					errors.New("get-env-var-group-error"),
				)
			})

			It("fetches the environment variable group from CC", func() {
				Expect(fakeCloudControllerClient.GetEnvironmentVariableGroupCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetEnvironmentVariableGroupArgsForCall(0)).To(Equal(constant.StagingEnvironmentVariableGroup))
			})

			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("get-env-var-group-error"))
				Expect(warnings).To(ConsistOf("get-env-var-group-warning"))
			})
		})

		When("getting the environment variable group succeeds", func() {
			var envVars ccv3.EnvironmentVariables

			BeforeEach(func() {
				envVars = map[string]types.FilteredString{
					"var_one": {IsSet: true, Value: "val_one"},
				}

				fakeCloudControllerClient.GetEnvironmentVariableGroupReturns(
					envVars,
					ccv3.Warnings{"get-env-var-group-warning"},
					nil,
				)
			})

			It("fetches the environment variable group from CC", func() {
				Expect(fakeCloudControllerClient.GetEnvironmentVariableGroupCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetEnvironmentVariableGroupArgsForCall(0)).To(Equal(constant.StagingEnvironmentVariableGroup))
			})

			It("returns result and warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fetchedEnvVariables).To(Equal(EnvironmentVariableGroup(envVars)))
				Expect(warnings).To(ConsistOf("get-env-var-group-warning"))
			})
		})
	})

	Describe("SetEnvironmentVariableGroup", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
			executeErr                error
			warnings                  Warnings
			envVars                   ccv3.EnvironmentVariables
		)

		BeforeEach(func() {
			envVars = ccv3.EnvironmentVariables{}
			fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil)
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.SetEnvironmentVariableGroup(constant.RunningEnvironmentVariableGroup, envVars)
		})

		When("Setting the environment variable group fails", func() {
			When("user passes some env var group", func() {
				BeforeEach(func() {
					envVars = ccv3.EnvironmentVariables{
						"key1": {Value: "val1", IsSet: true},
						"key2": {Value: "val2", IsSet: true},
					}

					fakeCloudControllerClient.UpdateEnvironmentVariableGroupReturns(
						nil,
						ccv3.Warnings{"update-env-var-group-warning"},
						errors.New("update-env-var-group-error"),
					)
				})

				It("sets the environment variable group via CC", func() {
					Expect(fakeCloudControllerClient.UpdateEnvironmentVariableGroupCallCount()).To(Equal(1))
					actualGroup, actualEnvPair := fakeCloudControllerClient.UpdateEnvironmentVariableGroupArgsForCall(0)
					Expect(actualGroup).To(Equal(constant.RunningEnvironmentVariableGroup))
					Expect(actualEnvPair).To(Equal(envVars))
				})

				It("returns warnings and error", func() {
					Expect(executeErr).To(MatchError("update-env-var-group-error"))
					Expect(warnings).To(ConsistOf("update-env-var-group-warning"))
				})

			})

			When("user passes in '{}'", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetEnvironmentVariableGroupReturns(
						ccv3.EnvironmentVariables{},
						ccv3.Warnings{},
						errors.New("I love my corgi, Pancho!"),
					)
				})

				It("propagates the error", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr).To(MatchError(errors.New("I love my corgi, Pancho!")))
				})
			})
		})

		When("Setting the environment variable group succeeds", func() {
			When("user passes some env var group", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateEnvironmentVariableGroupReturns(
						ccv3.EnvironmentVariables{
							"key1": {Value: "val1", IsSet: true},
							"key2": {Value: "val2", IsSet: true},
						},
						ccv3.Warnings{"update-env-var-group-warning"},
						nil,
					)
				})

				It("makes the API call to update the environment variable group and returns all warnings", func() {
					Expect(fakeCloudControllerClient.UpdateEnvironmentVariableGroupCallCount()).To(Equal(1))
					actualGroup, actualEnvPair := fakeCloudControllerClient.UpdateEnvironmentVariableGroupArgsForCall(0)
					Expect(actualGroup).To(Equal(constant.RunningEnvironmentVariableGroup))
					Expect(actualEnvPair).To(Equal(envVars))
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("update-env-var-group-warning"))
				})
			})

			When("user passes in '{}'", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetEnvironmentVariableGroupReturns(
						ccv3.EnvironmentVariables{
							"delete-me1": {Value: "val1", IsSet: true},
							"delete-me2": {Value: "val2", IsSet: true},
						},
						ccv3.Warnings{"get-env-var-group-warning"},
						nil,
					)
					fakeCloudControllerClient.UpdateEnvironmentVariableGroupReturns(
						ccv3.EnvironmentVariables{},
						ccv3.Warnings{"update-env-var-group-warning"},
						nil,
					)
				})

				It("nils the values of existing vars", func() {
					Expect(fakeCloudControllerClient.GetEnvironmentVariableGroupCallCount()).To(Equal(1))
					actualGroup := fakeCloudControllerClient.GetEnvironmentVariableGroupArgsForCall(0)
					Expect(actualGroup).To(Equal(constant.RunningEnvironmentVariableGroup))

					actualGroup, actualEnvPair := fakeCloudControllerClient.UpdateEnvironmentVariableGroupArgsForCall(0)
					Expect(actualGroup).To(Equal(constant.RunningEnvironmentVariableGroup))

					Expect(actualEnvPair).To(Equal(ccv3.EnvironmentVariables{
						"delete-me1": {Value: "", IsSet: false},
						"delete-me2": {Value: "", IsSet: false},
					}))

					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-env-var-group-warning", "update-env-var-group-warning"))
				})
			})

			When("user excludes some existing env vars", func() {
				BeforeEach(func() {
					envVars = ccv3.EnvironmentVariables{
						"keep-me": {Value: "val1", IsSet: true},
					}
					fakeCloudControllerClient.GetEnvironmentVariableGroupReturns(
						ccv3.EnvironmentVariables{
							"keep-me":   {Value: "val1", IsSet: true},
							"delete-me": {Value: "val2", IsSet: true},
						},
						ccv3.Warnings{"get-env-var-group-warning"},
						nil,
					)
					fakeCloudControllerClient.UpdateEnvironmentVariableGroupReturns(
						envVars,
						ccv3.Warnings{"update-env-var-group-warning"},
						nil,
					)
				})

				It("nils the values of excluded existing vars", func() {
					Expect(fakeCloudControllerClient.GetEnvironmentVariableGroupCallCount()).To(Equal(1))
					actualGroup := fakeCloudControllerClient.GetEnvironmentVariableGroupArgsForCall(0)
					Expect(actualGroup).To(Equal(constant.RunningEnvironmentVariableGroup))

					actualGroup, actualEnvPair := fakeCloudControllerClient.UpdateEnvironmentVariableGroupArgsForCall(0)
					Expect(actualGroup).To(Equal(constant.RunningEnvironmentVariableGroup))
					Expect(actualEnvPair).To(Equal(ccv3.EnvironmentVariables{
						"keep-me":   {Value: "val1", IsSet: true},
						"delete-me": {Value: "", IsSet: false},
					}))

					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-env-var-group-warning", "update-env-var-group-warning"))
				})
			})
		})
	})

	Describe("GetEnvironmentVariablesByApplicationNameAndSpace", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
			appName                   string
			spaceGUID                 string
			fetchedEnvVariables       EnvironmentVariableGroups
			executeErr                error
			warnings                  Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil)
			appName = "some-app"
			spaceGUID = "space-guid"
		})

		JustBeforeEach(func() {
			fetchedEnvVariables, warnings, executeErr = actor.GetEnvironmentVariablesByApplicationNameAndSpace(appName, spaceGUID)
		})

		When("finding the app fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(nil, ccv3.Warnings{"get-application-warning"}, errors.New("get-application-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("get-application-error"))
				Expect(warnings).To(ConsistOf("get-application-warning"))
			})
		})

		When("finding the app succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]resources.Application{{Name: "some-app", GUID: "some-app-guid"}}, ccv3.Warnings{"get-application-warning"}, nil)
			})

			When("getting the app environment variables fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationEnvironmentReturns(ccv3.Environment{}, ccv3.Warnings{"some-env-var-warnings"}, errors.New("some-env-var-error"))
				})
				It("returns an error", func() {
					Expect(executeErr).To(MatchError("some-env-var-error"))
					Expect(warnings).To(ConsistOf("get-application-warning", "some-env-var-warnings"))
				})
			})

			When("getting the app environment variables succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationEnvironmentReturns(
						ccv3.Environment{
							System:               map[string]interface{}{"system-var": "system-val"},
							Application:          map[string]interface{}{"app-var": "app-val"},
							EnvironmentVariables: map[string]interface{}{"user-var": "user-val"},
							Running:              map[string]interface{}{"running-var": "running-val"},
							Staging:              map[string]interface{}{"staging-var": "staging-val"},
						},
						ccv3.Warnings{"some-env-var-warnings"},
						nil,
					)
				})

				It("makes the API call to get the app environment variables and returns all warnings", func() {
					Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{appName}},
						ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
					))

					Expect(fakeCloudControllerClient.GetApplicationEnvironmentCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationEnvironmentArgsForCall(0)).To(Equal("some-app-guid"))
					Expect(fetchedEnvVariables).To(Equal(EnvironmentVariableGroups{
						System:               map[string]interface{}{"system-var": "system-val"},
						Application:          map[string]interface{}{"app-var": "app-val"},
						EnvironmentVariables: map[string]interface{}{"user-var": "user-val"},
						Running:              map[string]interface{}{"running-var": "running-val"},
						Staging:              map[string]interface{}{"staging-var": "staging-val"},
					}))
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-application-warning", "some-env-var-warnings"))
				})
			})
		})
	})

	Describe("SetEnvironmentVariableByApplicationNameAndSpace", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
			appName                   string
			spaceGUID                 string
			envPair                   EnvironmentVariablePair
			executeErr                error
			warnings                  Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil)
			appName = "some-app"
			spaceGUID = "space-guid"
			envPair = EnvironmentVariablePair{Key: "my-var", Value: "my-val"}
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.SetEnvironmentVariableByApplicationNameAndSpace(appName, spaceGUID, envPair)
		})

		When("finding the app fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(nil, ccv3.Warnings{"get-application-warning"}, errors.New("get-application-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("get-application-error"))
				Expect(warnings).To(ConsistOf("get-application-warning"))
			})
		})

		When("finding the app succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]resources.Application{{Name: "some-app", GUID: "some-app-guid"}}, ccv3.Warnings{"get-application-warning"}, nil)
			})

			When("updating the app environment variables fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateApplicationEnvironmentVariablesReturns(ccv3.EnvironmentVariables{}, ccv3.Warnings{"some-env-var-warnings"}, errors.New("some-env-var-error"))
				})
				It("returns an error", func() {
					Expect(executeErr).To(MatchError("some-env-var-error"))
					Expect(warnings).To(ConsistOf("get-application-warning", "some-env-var-warnings"))
				})
			})

			When("updating the app environment variables succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateApplicationEnvironmentVariablesReturns(
						ccv3.EnvironmentVariables{
							"my-var": {Value: "my-val", IsSet: true},
						},
						ccv3.Warnings{"some-env-var-warnings"},
						nil,
					)
				})

				It("makes the API call to update the app environment variables and returns all warnings", func() {
					Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{appName}},
						ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
					))

					Expect(fakeCloudControllerClient.UpdateApplicationEnvironmentVariablesCallCount()).To(Equal(1))
					appGUIDArg, envVarsArg := fakeCloudControllerClient.UpdateApplicationEnvironmentVariablesArgsForCall(0)
					Expect(appGUIDArg).To(Equal("some-app-guid"))
					Expect(envVarsArg).To(Equal(ccv3.EnvironmentVariables{
						"my-var": {Value: "my-val", IsSet: true},
					}))
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-application-warning", "some-env-var-warnings"))
				})
			})
		})
	})

	Describe("UnsetEnvironmentVariableByApplicationNameAndSpace", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
			appName                   string
			spaceGUID                 string
			envVariableName           string
			executeErr                error
			warnings                  Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil)
			appName = "some-app"
			spaceGUID = "space-guid"
			envVariableName = "my-var"
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.UnsetEnvironmentVariableByApplicationNameAndSpace(appName, spaceGUID, envVariableName)
		})

		When("finding the app fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(nil, ccv3.Warnings{"get-application-warning"}, errors.New("get-application-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("get-application-error"))
				Expect(warnings).To(ConsistOf("get-application-warning"))
			})
		})

		When("finding the app succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]resources.Application{{Name: "some-app", GUID: "some-app-guid"}}, ccv3.Warnings{"get-application-warning"}, nil)
			})

			When("getting the app environment variables fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationEnvironmentReturns(ccv3.Environment{}, ccv3.Warnings{"some-get-env-var-warnings"}, errors.New("some-env-var-error"))
				})
				It("returns an error", func() {
					Expect(executeErr).To(MatchError("some-env-var-error"))
					Expect(warnings).To(ConsistOf("get-application-warning", "some-get-env-var-warnings"))
				})
			})

			When("the variable doesn't exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationEnvironmentReturns(
						ccv3.Environment{},
						ccv3.Warnings{"some-get-env-var-warnings"},
						nil,
					)
				})
				It("returns an EnvironmentVariableNotSetError", func() {
					Expect(executeErr).To(MatchError(actionerror.EnvironmentVariableNotSetError{EnvironmentVariableName: "my-var"}))
					Expect(warnings).To(ConsistOf("get-application-warning", "some-get-env-var-warnings"))
				})
			})

			When("the variable exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationEnvironmentReturns(
						ccv3.Environment{
							EnvironmentVariables: map[string]interface{}{"my-var": "my-val"},
						},
						ccv3.Warnings{"some-get-env-var-warnings"},
						nil,
					)
				})
				When("updating the app environment variables fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.UpdateApplicationEnvironmentVariablesReturns(ccv3.EnvironmentVariables{}, ccv3.Warnings{"some-patch-env-var-warnings"}, errors.New("some-env-var-error"))
					})
					It("returns an error", func() {
						Expect(executeErr).To(MatchError("some-env-var-error"))
						Expect(warnings).To(ConsistOf("get-application-warning", "some-get-env-var-warnings", "some-patch-env-var-warnings"))
					})
				})

				When("updating the app environment variables succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.UpdateApplicationEnvironmentVariablesReturns(
							ccv3.EnvironmentVariables{
								"my-var": {Value: "my-val", IsSet: true},
							},
							ccv3.Warnings{"some-patch-env-var-warnings"},
							nil,
						)
					})

					It("makes the API call to update the app environment variables and returns all warnings", func() {
						Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
							ccv3.Query{Key: ccv3.NameFilter, Values: []string{appName}},
							ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
						))

						Expect(fakeCloudControllerClient.UpdateApplicationEnvironmentVariablesCallCount()).To(Equal(1))
						appGUIDArg, envVarsArg := fakeCloudControllerClient.UpdateApplicationEnvironmentVariablesArgsForCall(0)
						Expect(appGUIDArg).To(Equal("some-app-guid"))
						Expect(envVarsArg).To(Equal(ccv3.EnvironmentVariables{
							"my-var": {Value: "", IsSet: false},
						}))

						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("get-application-warning", "some-get-env-var-warnings", "some-patch-env-var-warnings"))
					})
				})
			})
		})
	})
})
