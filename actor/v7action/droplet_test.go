package v7action_test

import (
	"errors"
	"io"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Droplet Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("CreateApplicationDroplet", func() {
		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateDropletReturns(
					resources.Droplet{
						GUID:      "some-droplet-guid",
						State:     constant.DropletAwaitingUpload,
						CreatedAt: "2017-08-14T21:16:42Z",
						Image:     "docker/some-image",
						Stack:     "penguin",
					},
					ccv3.Warnings{"create-application-droplet-warning"},
					nil,
				)
			})

			It("creates a droplet for the app", func() {
				droplet, warnings, err := actor.CreateApplicationDroplet("some-app-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("create-application-droplet-warning"))
				Expect(droplet).To(Equal(resources.Droplet{
					GUID:      "some-droplet-guid",
					State:     constant.DropletAwaitingUpload,
					CreatedAt: "2017-08-14T21:16:42Z",
					Image:     "docker/some-image",
					Stack:     "penguin",
				}))

				Expect(fakeCloudControllerClient.CreateDropletCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateDropletArgsForCall(0)).To(Equal("some-app-guid"))
			})
		})

		When("creating the application droplet fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some upload droplet error")

				fakeCloudControllerClient.CreateDropletReturns(
					resources.Droplet{},
					ccv3.Warnings{"create-application-droplet-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.CreateApplicationDroplet("some-app-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("create-application-droplet-warning"))
			})
		})
	})

	Describe("SetApplicationDropletByApplicationNameAndSpace", func() {
		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.SetApplicationDropletReturns(
					resources.Relationship{GUID: "some-droplet-guid"},
					ccv3.Warnings{"set-application-droplet-warning"},
					nil,
				)
			})

			It("sets the app's droplet", func() {
				warnings, err := actor.SetApplicationDropletByApplicationNameAndSpace("some-app-name", "some-space-guid", "some-droplet-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-applications-warning", "set-application-droplet-warning"))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app-name"}},
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
				))

				Expect(fakeCloudControllerClient.SetApplicationDropletCallCount()).To(Equal(1))
				appGUID, dropletGUID := fakeCloudControllerClient.SetApplicationDropletArgsForCall(0)
				Expect(appGUID).To(Equal("some-app-guid"))
				Expect(dropletGUID).To(Equal("some-droplet-guid"))
			})
		})

		When("getting the application fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some get application error")

				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{},
					ccv3.Warnings{"get-applications-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				warnings, err := actor.SetApplicationDropletByApplicationNameAndSpace("some-app-name", "some-space-guid", "some-droplet-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-applications-warning"))
			})
		})

		When("setting the droplet fails", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("some set application-droplet error")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.SetApplicationDropletReturns(
					resources.Relationship{},
					ccv3.Warnings{"set-application-droplet-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				warnings, err := actor.SetApplicationDropletByApplicationNameAndSpace("some-app-name", "some-space-guid", "some-droplet-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-applications-warning", "set-application-droplet-warning"))
			})

			When("the cc client response contains an UnprocessableEntityError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.SetApplicationDropletReturns(
						resources.Relationship{},
						ccv3.Warnings{"set-application-droplet-warning"},
						ccerror.UnprocessableEntityError{Message: "some-message"},
					)
				})

				It("raises the error as AssignDropletError and returns warnings", func() {
					warnings, err := actor.SetApplicationDropletByApplicationNameAndSpace("some-app-name", "some-space-guid", "some-droplet-guid")

					Expect(err).To(MatchError("some-message"))
					Expect(warnings).To(ConsistOf("get-applications-warning", "set-application-droplet-warning"))
				})
			})

		})
	})

	Describe("SetApplicationDroplet", func() {
		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.SetApplicationDropletReturns(
					resources.Relationship{GUID: "some-droplet-guid"},
					ccv3.Warnings{"set-application-droplet-warning"},
					nil,
				)
			})

			It("sets the app's droplet", func() {
				warnings, err := actor.SetApplicationDroplet("some-app-guid", "some-droplet-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("set-application-droplet-warning"))

				Expect(fakeCloudControllerClient.SetApplicationDropletCallCount()).To(Equal(1))
				appGUID, dropletGUID := fakeCloudControllerClient.SetApplicationDropletArgsForCall(0)
				Expect(appGUID).To(Equal("some-app-guid"))
				Expect(dropletGUID).To(Equal("some-droplet-guid"))
			})
		})

		When("setting the droplet fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some set application-droplet error")

				fakeCloudControllerClient.SetApplicationDropletReturns(
					resources.Relationship{},
					ccv3.Warnings{"set-application-droplet-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				warnings, err := actor.SetApplicationDroplet("some-app-guid", "some-droplet-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("set-application-droplet-warning"))
			})

			When("the cc client response contains an UnprocessableEntityError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.SetApplicationDropletReturns(
						resources.Relationship{},
						ccv3.Warnings{"set-application-droplet-warning"},
						ccerror.UnprocessableEntityError{Message: "some-message"},
					)
				})

				It("raises the error as AssignDropletError and returns warnings", func() {
					warnings, err := actor.SetApplicationDroplet("some-app-guid", "some-droplet-guid")

					Expect(err).To(MatchError("some-message"))
					Expect(warnings).To(ConsistOf("set-application-droplet-warning"))
				})
			})
		})
	})

	Describe("GetApplicationDroplets", func() {
		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.GetDropletsReturns(
					[]resources.Droplet{
						{
							GUID:      "some-droplet-guid-1",
							State:     constant.DropletStaged,
							CreatedAt: "2017-08-14T21:16:42Z",
							Buildpacks: []resources.DropletBuildpack{
								{Name: "ruby"},
								{Name: "nodejs"},
							},
							Image: "docker/some-image",
							Stack: "penguin",
						},
						{
							GUID:      "some-droplet-guid-2",
							State:     constant.DropletFailed,
							CreatedAt: "2017-08-16T00:18:24Z",
							Buildpacks: []resources.DropletBuildpack{
								{Name: "java"},
							},
							Stack: "windows",
						},
					},
					ccv3.Warnings{"get-application-droplets-warning"},
					nil,
				)

				fakeCloudControllerClient.GetApplicationDropletCurrentReturns(
					resources.Droplet{
						GUID:      "some-droplet-guid-2",
						State:     constant.DropletFailed,
						CreatedAt: "2017-08-16T00:18:24Z",
						Buildpacks: []resources.DropletBuildpack{
							{Name: "java"},
						},
						Stack: "windows",
					},
					ccv3.Warnings{"get-current-droplet-warning"},
					nil,
				)
			})

			It("gets the app's droplets", func() {
				droplets, warnings, err := actor.GetApplicationDroplets("some-app-name", "some-space-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-applications-warning", "get-application-droplets-warning", "get-current-droplet-warning"))
				Expect(droplets).To(Equal([]resources.Droplet{
					{
						GUID:      "some-droplet-guid-1",
						State:     constant.DropletStaged,
						CreatedAt: "2017-08-14T21:16:42Z",
						Buildpacks: []resources.DropletBuildpack{
							{Name: "ruby"},
							{Name: "nodejs"},
						},
						Image:     "docker/some-image",
						Stack:     "penguin",
						IsCurrent: false,
					},
					{
						GUID:      "some-droplet-guid-2",
						State:     constant.DropletFailed,
						CreatedAt: "2017-08-16T00:18:24Z",
						Buildpacks: []resources.DropletBuildpack{
							{Name: "java"},
						},
						Stack:     "windows",
						IsCurrent: true,
					},
				}))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app-name"}},
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
				))

				Expect(fakeCloudControllerClient.GetDropletsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetDropletsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{"some-app-guid"}},
					ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.CreatedAtDescendingOrder}},
				))

				Expect(fakeCloudControllerClient.GetApplicationDropletCurrentCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationDropletCurrentArgsForCall(0)).To(Equal("some-app-guid"))
			})
		})

		When("the application does not have associated droplets", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.GetDropletsReturns(
					[]resources.Droplet{},
					ccv3.Warnings{"get-application-droplets-warning"},
					nil,
				)
			})

			It("does not error", func() {
				_, warnings, err := actor.GetApplicationDroplets("some-app-name", "some-space-guid")

				Expect(err).To(BeNil())
				Expect(warnings).To(ConsistOf("get-applications-warning", "get-application-droplets-warning"))

				Expect(fakeCloudControllerClient.GetApplicationDropletCurrentCallCount()).To(Equal(0))
			})
		})

		When("getting the application fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some get application error")

				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{},
					ccv3.Warnings{"get-applications-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.GetApplicationDroplets("some-app-name", "some-space-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-applications-warning"))
			})
		})

		When("getting the application droplets fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some get application error")

				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.GetDropletsReturns(
					[]resources.Droplet{},
					ccv3.Warnings{"get-application-droplets-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.GetApplicationDroplets("some-app-name", "some-space-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-applications-warning", "get-application-droplets-warning"))
			})
		})

		When("the application does not have a current droplet set", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.GetDropletsReturns(
					[]resources.Droplet{
						{
							GUID:      "some-droplet-guid-1",
							State:     constant.DropletStaged,
							CreatedAt: "2017-08-14T21:16:42Z",
							Buildpacks: []resources.DropletBuildpack{
								{Name: "ruby"},
								{Name: "nodejs"},
							},
							Image: "docker/some-image",
							Stack: "penguin",
						},
					},
					ccv3.Warnings{"get-application-droplets-warning"},
					nil,
				)

				fakeCloudControllerClient.GetApplicationDropletCurrentReturns(
					resources.Droplet{},
					ccv3.Warnings{"get-current-droplet-warning"},
					ccerror.DropletNotFoundError{},
				)
			})

			It("does not error and returns all droplets", func() {
				droplets, warnings, err := actor.GetApplicationDroplets("some-app-name", "some-space-guid")

				Expect(err).To(Not(HaveOccurred()))
				Expect(warnings).To(ConsistOf("get-applications-warning", "get-application-droplets-warning", "get-current-droplet-warning"))
				Expect(droplets).To(Equal([]resources.Droplet{
					{
						GUID:      "some-droplet-guid-1",
						State:     constant.DropletStaged,
						CreatedAt: "2017-08-14T21:16:42Z",
						Buildpacks: []resources.DropletBuildpack{
							{Name: "ruby"},
							{Name: "nodejs"},
						},
						Image:     "docker/some-image",
						Stack:     "penguin",
						IsCurrent: false,
					},
				}))
			})
		})
	})

	Describe("GetCurrentDropletByApplication", func() {
		var (
			appGUID string

			currentDroplet resources.Droplet
			warnings       Warnings
			executionErr   error
		)

		BeforeEach(func() {
			appGUID = "some-app-guid"
		})

		JustBeforeEach(func() {
			currentDroplet, warnings, executionErr = actor.GetCurrentDropletByApplication(appGUID)
		})

		When("the current droplet exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationDropletCurrentReturns(resources.Droplet{GUID: "some-droplet-guid"}, ccv3.Warnings{"some-warning"}, nil)
			})

			It("returns the current droplet", func() {
				Expect(executionErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(currentDroplet).To(Equal(resources.Droplet{GUID: "some-droplet-guid"}))

				Expect(fakeCloudControllerClient.GetApplicationDropletCurrentCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationDropletCurrentArgsForCall(0)).To(Equal("some-app-guid"))
			})
		})

		When("an error occurs", func() {
			When("the app does not exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationDropletCurrentReturns(resources.Droplet{GUID: "some-droplet-guid"}, ccv3.Warnings{"some-warning"}, ccerror.ApplicationNotFoundError{})
				})

				It("returns an ApplicationNotFoundError and warnings", func() {
					Expect(executionErr).To(MatchError(actionerror.ApplicationNotFoundError{GUID: appGUID}))
					Expect(warnings).To(ConsistOf("some-warning"))
				})
			})

			When("the current droplet does not exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationDropletCurrentReturns(resources.Droplet{}, ccv3.Warnings{"some-warning"}, ccerror.DropletNotFoundError{})
				})

				It("returns an DropletNotFoundError and warnings", func() {
					Expect(executionErr).To(MatchError(actionerror.DropletNotFoundError{AppGUID: appGUID}))
					Expect(warnings).To(ConsistOf("some-warning"))
				})
			})

			When("a generic error occurs", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some error")
					fakeCloudControllerClient.GetApplicationDropletCurrentReturns(resources.Droplet{}, ccv3.Warnings{"some-warning"}, expectedErr)
				})

				It("returns the error and warnings", func() {
					Expect(executionErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("some-warning"))
				})
			})
		})
	})

	Describe("UploadDroplet", func() {
		var (
			dropletGUID     string
			dropletFilePath string
			reader          io.Reader
			readerLength    int64

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			dropletGUID = "some-droplet-guid"
			dropletContents := "i am a droplet"
			reader = strings.NewReader(dropletContents)
			readerLength = int64(len([]byte(dropletContents)))
		})

		JustBeforeEach(func() {
			dropletFilePath = "tmp/droplet.tgz"
			warnings, executeErr = actor.UploadDroplet(dropletGUID, dropletFilePath, reader, readerLength)
		})

		When("the upload is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadDropletBitsReturns(
					ccv3.JobURL("some-job-url"),
					ccv3.Warnings{"some-upload-warning"},
					nil,
				)

				fakeCloudControllerClient.PollJobReturns(
					ccv3.Warnings{"some-job-warning"},
					nil,
				)
			})

			It("returns any CC warnings", func() {
				Expect(warnings).To(Equal(Warnings{"some-upload-warning", "some-job-warning"}))
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})

		When("the upload returns an error", func() {
			var uploadErr = errors.New("upload failed")

			BeforeEach(func() {
				fakeCloudControllerClient.UploadDropletBitsReturns(
					"",
					ccv3.Warnings{"some-upload-warning"},
					uploadErr,
				)
			})

			It("returns any warnings and the error", func() {
				Expect(warnings).To(Equal(Warnings{"some-upload-warning"}))
				Expect(executeErr).To(Equal(uploadErr))
			})
		})

		When("the upload job fails eventually", func() {
			var jobErr = errors.New("job failed")

			BeforeEach(func() {
				fakeCloudControllerClient.UploadDropletBitsReturns(
					ccv3.JobURL("some-job-url"),
					ccv3.Warnings{"some-upload-warning"},
					nil,
				)

				fakeCloudControllerClient.PollJobReturns(
					ccv3.Warnings{"some-job-warning"},
					jobErr,
				)
			})

			It("returns any warning and the error", func() {
				Expect(warnings).To(Equal(Warnings{"some-upload-warning", "some-job-warning"}))
				Expect(executeErr).To(Equal(jobErr))
			})
		})
	})
})
