package v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Environment Variable Actions", func() {
	Describe("GetEnvironmentVariablesByApplicationNameAndSpace", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
			appName                   string
			spaceGUID                 string
			fetchedEnvVariables       EnvironmentVariableGroups
			executeErr                error
			warnings                  Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
			appName = "some-app"
			spaceGUID = "space-guid"
		})

		JustBeforeEach(func() {
			fetchedEnvVariables, warnings, executeErr = actor.GetEnvironmentVariablesByApplicationNameAndSpace(appName, spaceGUID)
		})

		Context("when finding the app fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(nil, ccv3.Warnings{"get-application-warning"}, errors.New("get-application-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("get-application-error"))
				Expect(warnings).To(ConsistOf("get-application-warning"))
			})
		})

		Context("when finding the app succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{Name: "some-app", GUID: "some-app-guid"}}, ccv3.Warnings{"get-application-warning"}, nil)
			})

			Context("when getting the app environment variables fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationEnvironmentVariablesReturns(ccv3.EnvironmentVariableGroups{}, ccv3.Warnings{"some-env-var-warnings"}, errors.New("some-env-var-error"))
				})
				It("returns an error", func() {
					Expect(executeErr).To(MatchError("some-env-var-error"))
					Expect(warnings).To(ConsistOf("get-application-warning", "some-env-var-warnings"))
				})
			})

			Context("when getting the app environment variables succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationEnvironmentVariablesReturns(
						ccv3.EnvironmentVariableGroups{
							SystemProvided:      map[string]interface{}{"system-var": "system-val"},
							ApplicationProvided: map[string]interface{}{"app-var": "app-val"},
							UserProvided:        map[string]interface{}{"user-var": "user-val"},
							RunningGroup:        map[string]interface{}{"running-var": "running-val"},
							StagingGroup:        map[string]interface{}{"staging-var": "staging-val"},
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

					Expect(fakeCloudControllerClient.GetApplicationEnvironmentVariablesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationEnvironmentVariablesArgsForCall(0)).To(Equal("some-app-guid"))
					Expect(fetchedEnvVariables).To(Equal(EnvironmentVariableGroups{
						SystemProvided:      map[string]interface{}{"system-var": "system-val"},
						ApplicationProvided: map[string]interface{}{"app-var": "app-val"},
						UserProvided:        map[string]interface{}{"user-var": "user-val"},
						RunningGroup:        map[string]interface{}{"running-var": "running-val"},
						StagingGroup:        map[string]interface{}{"staging-var": "staging-val"},
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
			fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
			appName                   string
			spaceGUID                 string
			envPair                   EnvironmentVariablePair
			executeErr                error
			warnings                  Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
			appName = "some-app"
			spaceGUID = "space-guid"
			envPair = EnvironmentVariablePair{Key: "my-var", Value: "my-val"}
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.SetEnvironmentVariableByApplicationNameAndSpace(appName, spaceGUID, envPair)
		})

		Context("when finding the app fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(nil, ccv3.Warnings{"get-application-warning"}, errors.New("get-application-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("get-application-error"))
				Expect(warnings).To(ConsistOf("get-application-warning"))
			})
		})

		Context("when finding the app succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{Name: "some-app", GUID: "some-app-guid"}}, ccv3.Warnings{"get-application-warning"}, nil)
			})

			Context("when updating the app environment variables fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.PatchApplicationUserProvidedEnvironmentVariablesReturns(ccv3.EnvironmentVariables{}, ccv3.Warnings{"some-env-var-warnings"}, errors.New("some-env-var-error"))
				})
				It("returns an error", func() {
					Expect(executeErr).To(MatchError("some-env-var-error"))
					Expect(warnings).To(ConsistOf("get-application-warning", "some-env-var-warnings"))
				})
			})

			Context("when updating the app environment variables succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.PatchApplicationUserProvidedEnvironmentVariablesReturns(
						ccv3.EnvironmentVariables{
							Variables: map[string]types.FilteredString{"my-var": {Value: "my-val", IsSet: true}},
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

					Expect(fakeCloudControllerClient.PatchApplicationUserProvidedEnvironmentVariablesCallCount()).To(Equal(1))
					appGUIDArg, envVarsArg := fakeCloudControllerClient.PatchApplicationUserProvidedEnvironmentVariablesArgsForCall(0)
					Expect(appGUIDArg).To(Equal("some-app-guid"))
					Expect(envVarsArg).To(Equal(ccv3.EnvironmentVariables{
						Variables: map[string]types.FilteredString{"my-var": {Value: "my-val", IsSet: true}},
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
			fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
			appName                   string
			spaceGUID                 string
			envVariableName           string
			executeErr                error
			warnings                  Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
			appName = "some-app"
			spaceGUID = "space-guid"
			envVariableName = "my-var"
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.UnsetEnvironmentVariableByApplicationNameAndSpace(appName, spaceGUID, envVariableName)
		})

		Context("when finding the app fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(nil, ccv3.Warnings{"get-application-warning"}, errors.New("get-application-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("get-application-error"))
				Expect(warnings).To(ConsistOf("get-application-warning"))
			})
		})

		Context("when finding the app succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{Name: "some-app", GUID: "some-app-guid"}}, ccv3.Warnings{"get-application-warning"}, nil)
			})

			Context("when getting the app environment variables fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationEnvironmentVariablesReturns(ccv3.EnvironmentVariableGroups{}, ccv3.Warnings{"some-get-env-var-warnings"}, errors.New("some-env-var-error"))
				})
				It("returns an error", func() {
					Expect(executeErr).To(MatchError("some-env-var-error"))
					Expect(warnings).To(ConsistOf("get-application-warning", "some-get-env-var-warnings"))
				})
			})

			Context("when the variable doesn't exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationEnvironmentVariablesReturns(
						ccv3.EnvironmentVariableGroups{},
						ccv3.Warnings{"some-get-env-var-warnings"},
						nil,
					)
				})
				It("returns an EnvironmentVariableNotSetError", func() {
					Expect(executeErr).To(MatchError(actionerror.EnvironmentVariableNotSetError{EnvironmentVariableName: "my-var"}))
					Expect(warnings).To(ConsistOf("get-application-warning", "some-get-env-var-warnings"))
				})
			})

			Context("when the variable exists", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationEnvironmentVariablesReturns(
						ccv3.EnvironmentVariableGroups{
							UserProvided: map[string]interface{}{"my-var": "my-val"},
						},
						ccv3.Warnings{"some-get-env-var-warnings"},
						nil,
					)
				})
				Context("when updating the app environment variables fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.PatchApplicationUserProvidedEnvironmentVariablesReturns(ccv3.EnvironmentVariables{}, ccv3.Warnings{"some-patch-env-var-warnings"}, errors.New("some-env-var-error"))
					})
					It("returns an error", func() {
						Expect(executeErr).To(MatchError("some-env-var-error"))
						Expect(warnings).To(ConsistOf("get-application-warning", "some-get-env-var-warnings", "some-patch-env-var-warnings"))
					})
				})

				Context("when updating the app environment variables succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.PatchApplicationUserProvidedEnvironmentVariablesReturns(
							ccv3.EnvironmentVariables{
								Variables: map[string]types.FilteredString{"my-var": {Value: "my-val", IsSet: true}},
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

						Expect(fakeCloudControllerClient.PatchApplicationUserProvidedEnvironmentVariablesCallCount()).To(Equal(1))
						appGUIDArg, envVarsArg := fakeCloudControllerClient.PatchApplicationUserProvidedEnvironmentVariablesArgsForCall(0)
						Expect(appGUIDArg).To(Equal("some-app-guid"))
						Expect(envVarsArg).To(Equal(ccv3.EnvironmentVariables{
							Variables: map[string]types.FilteredString{"my-var": {Value: "", IsSet: false}},
						}))

						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("get-application-warning", "some-get-env-var-warnings", "some-patch-env-var-warnings"))
					})
				})
			})
		})
	})
})
