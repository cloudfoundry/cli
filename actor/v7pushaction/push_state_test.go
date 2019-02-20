package v7pushaction_test

import (
	"errors"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Push State", func() {
	var (
		actor           *Actor
		fakeV7Actor     *v7pushactionfakes.FakeV7Actor
		fakeSharedActor *v7pushactionfakes.FakeSharedActor

		pwd string
	)

	BeforeEach(func() {
		actor, _, fakeV7Actor, fakeSharedActor = getTestPushActor()
	})

	Describe("Conceptualize", func() {
		var (
			appNames      []string
			spaceGUID     string
			orgGUID       string
			currentDir    string
			flagOverrides FlagOverrides

			states     []PushState
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			var err error
			pwd, err = os.Getwd()
			Expect(err).ToNot(HaveOccurred())
			appNames = []string{"some-app-name"}
			currentDir = pwd
			flagOverrides = FlagOverrides{}

			spaceGUID = "some-space-guid"
			orgGUID = "some-org-guid"
		})

		JustBeforeEach(func() {
			states, warnings, executeErr = actor.Conceptualize(appNames, spaceGUID, orgGUID, currentDir, flagOverrides)
		})

		Describe("application", func() {
			When("the application exists", func() {
				var apps []v7action.Application

				BeforeEach(func() {
					apps = []v7action.Application{
						{
							GUID: "app-guid-0",
							Name: "app-name-0",
						},
						{
							GUID: "app-guid-1",
							Name: "app-name-1",
						},
						{
							GUID: "app-guid-2",
							Name: "app-name-2",
						},
					}
					fakeV7Actor.GetApplicationsByNamesAndSpaceReturns(apps, v7action.Warnings{"plural-get-warning"}, nil)
				})

				It("returns multiple push states", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("plural-get-warning"))
					Expect(states).To(HaveLen(len(apps)))

					Expect(states).To(ConsistOf(
						MatchFields(IgnoreExtras, Fields{
							"Application": Equal(apps[0]),
							"SpaceGUID":   Equal(spaceGUID),
							"OrgGUID":     Equal(orgGUID),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Application": Equal(apps[1]),
							"SpaceGUID":   Equal(spaceGUID),
							"OrgGUID":     Equal(orgGUID),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Application": Equal(apps[2]),
							"SpaceGUID":   Equal(spaceGUID),
							"OrgGUID":     Equal(orgGUID),
						}),
					))

					Expect(fakeV7Actor.GetApplicationsByNamesAndSpaceCallCount()).To(Equal(1))
				})
			})

			When("getting the app errors", func() {
				BeforeEach(func() {
					fakeV7Actor.GetApplicationsByNamesAndSpaceReturns([]v7action.Application{}, v7action.Warnings{"some-app-warning"}, errors.New("bad-get-error"))
				})
				It("errors", func() {
					Expect(executeErr).To(MatchError(errors.New("bad-get-error")))
					Expect(warnings).To(ConsistOf("some-app-warning"))
				})
			})
		})

		Describe("Push state construction", func() {
			var apps = []v7action.Application{
				{
					GUID: "app-guid-0",
					Name: "app-name-0",
				},
				{
					GUID: "app-guid-1",
					Name: "app-name-1",
				},
				{
					GUID: "app-guid-2",
					Name: "app-name-2",
				},
			}

			BeforeEach(func() {
				fakeV7Actor.GetApplicationsByNamesAndSpaceReturns(apps, v7action.Warnings{"plural-get-warning"}, nil)
			})

			It("defaults to needsUpdate = false", func() {
				Expect(states[0].ApplicationNeedsUpdate).To(Equal(false))
			})

			Describe("lifecycle types", func() {
				When("buildpacks are provided via flagOverrides", func() {
					BeforeEach(func() {
						flagOverrides.Buildpacks = []string{"some-buildpack-1", "some-buildpack-2"}
					})

					It("sets the buildpacks on the apps", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(states[0].Application.LifecycleType).To(Equal(constant.AppLifecycleTypeBuildpack))
						Expect(states[1].Application.LifecycleType).To(Equal(constant.AppLifecycleTypeBuildpack))
						Expect(states[0].Application.LifecycleBuildpacks).To(ConsistOf("some-buildpack-1", "some-buildpack-2"))
					})

					It("sets needsUpdate", func() {
						Expect(states[0].ApplicationNeedsUpdate).To(Equal(true))
					})
				})

				When("stack is provided via flagOverrides", func() {
					BeforeEach(func() {
						flagOverrides.Stack = "validStack"
					})

					It("sets the stack on the app", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(states[0].Application.LifecycleType).To(Equal(constant.AppLifecycleTypeBuildpack))
						Expect(states[0].Application.StackName).To(Equal("validStack"))
					})

					It("sets needsUpdate", func() {
						Expect(states[0].ApplicationNeedsUpdate).To(Equal(true))
					})
				})

				When("docker image information is provided", func() {
					BeforeEach(func() {
						flagOverrides.DockerImage = "some-docker-image"
						flagOverrides.DockerPassword = "some-docker-password"
						flagOverrides.DockerUsername = "some-docker-username"
					})

					It("sets the lifecycle info on the apps", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(states[0].Application.LifecycleType).To(Equal(constant.AppLifecycleTypeDocker))
						Expect(states[1].Application.LifecycleType).To(Equal(constant.AppLifecycleTypeDocker))
						Expect(states[0].Application.LifecycleBuildpacks).To(BeEmpty())
					})

					It("sets needsUpdate", func() {
						Expect(states[0].ApplicationNeedsUpdate).To(Equal(true))
					})
				})
			})

			Describe("bits path", func() {
				When("no app path is provided in the command line settings", func() {
					It("sets the bits path to the current directory in the settings", func() {
						Expect(states[0].BitsPath).To(Equal(pwd))
					})

					It("does not set needsUpdate", func() {
						Expect(states[0].ApplicationNeedsUpdate).To(Equal(false))
					})
				})

				When("an app path is provided in the command line settings", func() {
					var providedPath string

					BeforeEach(func() {
						archive, err := ioutil.TempFile("", "push-state-provided-path")
						Expect(err).ToNot(HaveOccurred())
						defer archive.Close()

						providedPath = archive.Name()
						flagOverrides.ProvidedAppPath = providedPath
					})

					It("sets the bits paths to the provided app path", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(states[0].BitsPath).To(Equal(providedPath))
						Expect(states[1].BitsPath).To(Equal(providedPath))
					})
				})
			})

			Describe("all resources/archive", func() {
				When("the app resources are given as a directory", func() {
					When("gathering the resources is successful", func() {
						var resources []sharedaction.Resource

						BeforeEach(func() {
							resources = []sharedaction.Resource{
								{
									Filename: "fake-app-file",
								},
							}
							fakeSharedActor.GatherDirectoryResourcesReturns(resources, nil)
						})

						It("adds the gathered resources to the push state", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(3))
							Expect(fakeSharedActor.GatherDirectoryResourcesArgsForCall(0)).To(Equal(pwd))
							Expect(states[0].AllResources).To(Equal(resources))
						})

						It("sets Archive to false", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(states[0].Archive).To(BeFalse())
						})
					})

					When("gathering the resources errors", func() {
						BeforeEach(func() {
							fakeSharedActor.GatherDirectoryResourcesReturns(nil, errors.New("kaboom"))
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError("kaboom"))
						})
					})
				})

				When("the app resources are given as an archive", func() {
					var archivePath string

					BeforeEach(func() {
						archive, err := ioutil.TempFile("", "push-state-archive")
						Expect(err).ToNot(HaveOccurred())
						defer archive.Close()

						archivePath = archive.Name()
						flagOverrides.ProvidedAppPath = archivePath
					})

					AfterEach(func() {
						Expect(os.RemoveAll(archivePath)).ToNot(HaveOccurred())
					})

					When("gathering the resources is successful", func() {
						var resources []sharedaction.Resource

						BeforeEach(func() {
							resources = []sharedaction.Resource{
								{
									Filename: "fake-app-file",
								},
							}
							fakeSharedActor.GatherArchiveResourcesReturns(resources, nil)
						})

						It("adds the gathered resources to the push state", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(fakeSharedActor.GatherArchiveResourcesCallCount()).To(Equal(3))
							Expect(fakeSharedActor.GatherArchiveResourcesArgsForCall(0)).To(Equal(archivePath))
							Expect(states[0].AllResources).To(Equal(resources))
						})

						It("sets Archive to true", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(states[0].Archive).To(BeTrue())
						})
					})

					When("gathering the resources errors", func() {
						BeforeEach(func() {
							fakeSharedActor.GatherArchiveResourcesReturns(nil, errors.New("kaboom"))
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError("kaboom"))
						})
					})
				})
			})

			When("flag overrides are passed", func() {
				BeforeEach(func() {
					flagOverrides.Memory = types.NullUint64{IsSet: true, Value: 123456}
				})

				It("sets the all the flag overrides on the state", func() {
					Expect(states[0].Overrides).To(Equal(flagOverrides))
				})
			})
		})
	})
})
