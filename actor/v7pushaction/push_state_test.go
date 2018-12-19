package v7pushaction_test

import (
	"errors"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
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
			appName       string
			spaceGUID     string
			orgGUID       string
			currentDir    string
			flagOverrides FlagOverrides
			manifest      []byte

			states     []PushState
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			var err error
			pwd, err = os.Getwd()
			Expect(err).ToNot(HaveOccurred())
			appName = "some-app-name"
			currentDir = pwd
			flagOverrides = FlagOverrides{}
			manifest = []byte("some yaml")

			spaceGUID = "some-space-guid"
			orgGUID = "some-org-guid"
		})

		JustBeforeEach(func() {
			states, warnings, executeErr = actor.Conceptualize(appName, spaceGUID, orgGUID, currentDir, flagOverrides, manifest)
		})

		Describe("application", func() {
			When("the application exists", func() {
				var app v7action.Application

				BeforeEach(func() {
					app = v7action.Application{
						GUID: "some-app-guid",
						Name: "some-app-name",
					}

					fakeV7Actor.GetApplicationByNameAndSpaceReturns(app, v7action.Warnings{"some-app-warning"}, nil)
				})

				It("uses the found app in the application state", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-app-warning"))
					Expect(states).To(HaveLen(1))

					Expect(states[0]).To(MatchFields(IgnoreExtras,
						Fields{
							"Application": Equal(app),
							"SpaceGUID":   Equal(spaceGUID),
							"OrgGUID":     Equal(orgGUID),
						}))

					Expect(fakeV7Actor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					passedName, passedSpaceGUID := fakeV7Actor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(passedName).To(Equal("some-app-name"))
					Expect(passedSpaceGUID).To(Equal(spaceGUID))
				})
			})

			When("the application does not exist", func() {
				BeforeEach(func() {
					fakeV7Actor.GetApplicationByNameAndSpaceReturns(v7action.Application{}, v7action.Warnings{"some-app-warning"}, actionerror.ApplicationNotFoundError{})
				})

				It("creates a new app in the application state", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-app-warning"))
					Expect(states).To(HaveLen(1))

					Expect(states[0]).To(MatchFields(IgnoreExtras,
						Fields{
							"Application": Equal(v7action.Application{
								Name: "some-app-name",
							}),
							"SpaceGUID": Equal(spaceGUID),
							"OrgGUID":   Equal(orgGUID),
						}))
				})
			})

			Context("regardless of application existance", func() {
				When("buildpacks are provided via flagOverrides", func() {
					BeforeEach(func() {
						flagOverrides.Buildpacks = []string{"some-buildpack-1", "some-buildpack-2"}
					})

					It("sets the buildpacks on the app", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(states[0].Application.LifecycleType).To(Equal(constant.AppLifecycleTypeBuildpack))
						Expect(states[0].Application.LifecycleBuildpacks).To(ConsistOf("some-buildpack-1", "some-buildpack-2"))
					})
				})

				When("docker image information is provided", func() {
					BeforeEach(func() {
						flagOverrides.DockerImage = "some-docker-image"
						flagOverrides.DockerPassword = "some-docker-password"
						flagOverrides.DockerUsername = "some-docker-username"
					})

					It("sets the buildpacks on the app", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(states[0].Application.LifecycleType).To(Equal(constant.AppLifecycleTypeDocker))
						Expect(states[0].Application.LifecycleBuildpacks).To(BeEmpty())
					})
				})
			})

			When("the application lookup errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeV7Actor.GetApplicationByNameAndSpaceReturns(v7action.Application{}, v7action.Warnings{"some-app-warning"}, expectedErr)
				})

				It("translates command line settings into a single push state", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("some-app-warning"))
				})
			})
		})

		Describe("bits path", func() {
			When("no app path is provided in the command line settings", func() {
				It("sets the bits path to the current directory in the settings", func() {
					Expect(states[0].BitsPath).To(Equal(pwd))
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

				It("sets the bits path to the provided app path", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(states[0].BitsPath).To(Equal(providedPath))
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
						Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(1))
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
						Expect(fakeSharedActor.GatherArchiveResourcesCallCount()).To(Equal(1))
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

		Describe("manifest", func() {
			It("attaches manifest to pushState", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(states[0]).To(MatchFields(IgnoreExtras,
					Fields{
						"Manifest": Equal(manifest),
					}))
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
