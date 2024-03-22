package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("CreateDockerPackageForApplication", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		returnedPushPlan PushPlan
		paramPlan        PushPlan
		fakeProgressBar  *v7pushactionfakes.FakeProgressBar

		warnings   Warnings
		executeErr error

		events []Event
	)

	BeforeEach(func() {
		actor, fakeV7Actor, _ = getTestPushActor()

		fakeProgressBar = new(v7pushactionfakes.FakeProgressBar)

		paramPlan = PushPlan{
			Application: resources.Application{
				GUID: "some-app-guid",
			},
			DockerImageCredentials: v7action.DockerImageCredentials{
				Path:     "some-docker-image",
				Password: "some-docker-password",
				Username: "some-docker-username",
			},
		}
	})

	JustBeforeEach(func() {
		events = EventFollower(func(eventStream chan<- *PushEvent) {
			returnedPushPlan, warnings, executeErr = actor.CreateDockerPackageForApplication(paramPlan, eventStream, fakeProgressBar)
		})
	})

	Describe("package upload", func() {
		BeforeEach(func() {
			fakeV7Actor.CreateApplicationInSpaceReturns(
				resources.Application{
					GUID:          "some-app-guid",
					Name:          paramPlan.Application.Name,
					LifecycleType: constant.AppLifecycleTypeDocker,
				},
				v7action.Warnings{"some-app-warnings"},
				nil)
		})

		When("creating the package is successful", func() {
			BeforeEach(func() {
				fakeV7Actor.CreateDockerPackageByApplicationReturns(
					resources.Package{GUID: "some-package-guid"},
					v7action.Warnings{"some-package-warnings"},
					nil)
			})

			It("sets the docker image", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-package-warnings"))
				Expect(events).To(ConsistOf(SetDockerImage, SetDockerImageComplete))
				Expect(fakeV7Actor.CreateDockerPackageByApplicationCallCount()).To(Equal(1))

				appGUID, dockerCredentials := fakeV7Actor.CreateDockerPackageByApplicationArgsForCall(0)
				Expect(appGUID).To(Equal("some-app-guid"))
				Expect(dockerCredentials).To(MatchFields(IgnoreExtras,
					Fields{
						"Path":     Equal("some-docker-image"),
						"Username": Equal("some-docker-username"),
						"Password": Equal("some-docker-password"),
					}))

				Expect(fakeV7Actor.PollPackageArgsForCall(0)).To(MatchFields(IgnoreExtras,
					Fields{
						"GUID": Equal("some-package-guid"),
					}))
			})
		})

		When("creating the package errors", func() {
			var someErr error

			BeforeEach(func() {
				someErr = errors.New("I AM A BANANA")
				fakeV7Actor.CreateDockerPackageByApplicationReturns(resources.Package{}, v7action.Warnings{"some-package-warnings"}, someErr)
			})

			It("returns errors and warnings", func() {
				Expect(executeErr).To(MatchError(someErr))
				Expect(events).To(ConsistOf(SetDockerImage))
				Expect(warnings).To(ConsistOf("some-package-warnings"))
			})
		})
	})

	Describe("polling package", func() {
		When("the the polling is successful", func() {
			BeforeEach(func() {
				fakeV7Actor.PollPackageReturns(resources.Package{GUID: "some-package-guid"}, v7action.Warnings{"some-poll-package-warning"}, nil)
			})

			It("returns warnings", func() {
				Expect(events).To(ConsistOf(SetDockerImage, SetDockerImageComplete))
				Expect(warnings).To(ConsistOf("some-poll-package-warning"))
			})

			It("sets the package guid on push plan", func() {
				Expect(returnedPushPlan.PackageGUID).To(Equal("some-package-guid"))
			})
		})

		When("the the polling returns an error", func() {
			var someErr error

			BeforeEach(func() {
				someErr = errors.New("I AM A BANANA")
				fakeV7Actor.PollPackageReturns(resources.Package{}, v7action.Warnings{"some-poll-package-warning"}, someErr)
			})

			It("returns errors and warnings", func() {
				Expect(events).To(ConsistOf(SetDockerImage, SetDockerImageComplete))
				Expect(executeErr).To(MatchError(someErr))
			})
		})
	})
})
