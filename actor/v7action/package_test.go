package v7action_test

import (
	"errors"
	"io"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Package Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		fakeSharedActor           *v7actionfakes.FakeSharedActor
		fakeConfig                *v7actionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v7actionfakes.FakeConfig)
		fakeSharedActor = new(v7actionfakes.FakeSharedActor)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, fakeSharedActor, nil)
	})

	Describe("GetApplicationPackages", func() {
		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.GetPackagesReturns(
					[]ccv3.Package{
						{
							GUID:      "some-package-guid-1",
							State:     constant.PackageReady,
							CreatedAt: "2017-08-14T21:16:42Z",
						},
						{
							GUID:      "some-package-guid-2",
							State:     constant.PackageFailed,
							CreatedAt: "2017-08-16T00:18:24Z",
						},
					},
					ccv3.Warnings{"get-application-packages-warning"},
					nil,
				)
			})

			It("gets the app's packages", func() {
				packages, warnings, err := actor.GetApplicationPackages("some-app-name", "some-space-guid")

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-applications-warning", "get-application-packages-warning"))
				Expect(packages).To(Equal([]Package{
					{
						GUID:      "some-package-guid-1",
						State:     constant.PackageReady,
						CreatedAt: "2017-08-14T21:16:42Z",
					},
					{
						GUID:      "some-package-guid-2",
						State:     constant.PackageFailed,
						CreatedAt: "2017-08-16T00:18:24Z",
					},
				}))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app-name"}},
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
				))

				Expect(fakeCloudControllerClient.GetPackagesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetPackagesArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{"some-app-guid"}},
				))
			})
		})

		When("getting the application fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some get application error")

				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"get-applications-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.GetApplicationPackages("some-app-name", "some-space-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-applications-warning"))
			})
		})

		When("getting the application packages fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some get application error")

				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.GetPackagesReturns(
					[]ccv3.Package{},
					ccv3.Warnings{"get-application-packages-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.GetApplicationPackages("some-app-name", "some-space-guid")

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-applications-warning", "get-application-packages-warning"))
			})
		})
	})

	Describe("CreateDockerPackageByApplicationNameAndSpace", func() {
		var (
			dockerPackage Package
			warnings      Warnings
			executeErr    error
		)

		JustBeforeEach(func() {
			dockerPackage, warnings, executeErr = actor.CreateDockerPackageByApplicationNameAndSpace("some-app-name", "some-space-guid", DockerImageCredentials{Path: "some-docker-image", Password: "some-password", Username: "some-username"})
		})

		When("the application can't be retrieved", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-app-warning"},
					errors.New("some-app-error"),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError("some-app-error"))
				Expect(warnings).To(ConsistOf("some-app-warning"))
			})
		})

		When("the application can be retrieved", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{
							Name: "some-app-name",
							GUID: "some-app-guid",
						},
					},
					ccv3.Warnings{"some-app-warning"},
					nil,
				)
			})

			When("creating the package fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreatePackageReturns(
						ccv3.Package{},
						ccv3.Warnings{"some-create-package-warning"},
						errors.New("some-create-package-error"),
					)
				})
				It("fails to create the package", func() {
					Expect(executeErr).To(MatchError("some-create-package-error"))
					Expect(warnings).To(ConsistOf("some-app-warning", "some-create-package-warning"))
				})
			})

			When("creating the package succeeds", func() {
				BeforeEach(func() {
					createdPackage := ccv3.Package{
						DockerImage:    "some-docker-image",
						DockerUsername: "some-username",
						DockerPassword: "some-password",
						GUID:           "some-pkg-guid",
						State:          constant.PackageReady,
						Relationships: ccv3.Relationships{
							constant.RelationshipTypeApplication: ccv3.Relationship{
								GUID: "some-app-guid",
							},
						},
					}

					fakeCloudControllerClient.CreatePackageReturns(
						createdPackage,
						ccv3.Warnings{"some-create-package-warning"},
						nil,
					)
				})

				It("calls CC to create the package and returns the package", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-app-warning", "some-create-package-warning"))

					expectedPackage := ccv3.Package{
						DockerImage:    "some-docker-image",
						DockerUsername: "some-username",
						DockerPassword: "some-password",
						GUID:           "some-pkg-guid",
						State:          constant.PackageReady,
						Relationships: ccv3.Relationships{
							constant.RelationshipTypeApplication: ccv3.Relationship{
								GUID: "some-app-guid",
							},
						},
					}
					Expect(dockerPackage).To(Equal(Package(expectedPackage)))

					Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app-name"}},
						ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
					))

					Expect(fakeCloudControllerClient.CreatePackageCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.CreatePackageArgsForCall(0)).To(Equal(ccv3.Package{
						Type:           constant.PackageTypeDocker,
						DockerImage:    "some-docker-image",
						DockerUsername: "some-username",
						DockerPassword: "some-password",
						Relationships: ccv3.Relationships{
							constant.RelationshipTypeApplication: ccv3.Relationship{GUID: "some-app-guid"},
						},
					}))
				})
			})
		})
	})

	Describe("CreateBitsPackageByApplication", func() {
		var (
			appGUID string

			pkg        Package
			executeErr error
			warnings   Warnings
		)

		JustBeforeEach(func() {
			pkg, warnings, executeErr = actor.CreateBitsPackageByApplication(appGUID)
		})

		When("creating the package fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreatePackageReturns(
					ccv3.Package{},
					ccv3.Warnings{"create-package-warning"},
					errors.New("some-create-error"),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("some-create-error"))
				Expect(warnings).To(ConsistOf("create-package-warning"))
			})
		})

		When("creating the package succeeds", func() {
			var createdPackage ccv3.Package

			BeforeEach(func() {
				createdPackage = ccv3.Package{GUID: "some-pkg-guid"}
				fakeCloudControllerClient.CreatePackageReturns(
					createdPackage,
					ccv3.Warnings{"create-package-warning"},
					nil,
				)
			})

			It("returns all warnings and the package", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.CreatePackageCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreatePackageArgsForCall(0)).To(Equal(ccv3.Package{
					Type: constant.PackageTypeBits,
					Relationships: ccv3.Relationships{
						constant.RelationshipTypeApplication: ccv3.Relationship{GUID: appGUID},
					},
				}))

				Expect(warnings).To(ConsistOf("create-package-warning"))
				Expect(pkg).To(MatchFields(IgnoreExtras, Fields{
					"GUID": Equal("some-pkg-guid"),
				}))
			})
		})
	})

	Describe("UploadBitsPackage", func() {
		var (
			pkg              Package
			matchedResources []sharedaction.V3Resource
			reader           io.Reader
			readerLength     int64

			appPkg     Package
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			pkg = Package{GUID: "some-package-guid"}

			matchedResources = []sharedaction.V3Resource{{FilePath: "some-resource"}, {FilePath: "another-resource"}}
			someString := "who reads these days"
			reader = strings.NewReader(someString)
			readerLength = int64(len([]byte(someString)))
		})

		JustBeforeEach(func() {
			appPkg, warnings, executeErr = actor.UploadBitsPackage(pkg, matchedResources, reader, readerLength)
		})

		When("the upload is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadBitsPackageReturns(ccv3.Package{GUID: "some-package-guid"}, ccv3.Warnings{"upload-warning-1", "upload-warning-2"}, nil)
			})

			It("returns all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2"))
				Expect(appPkg).To(Equal(Package{GUID: "some-package-guid"}))

				Expect(fakeCloudControllerClient.UploadBitsPackageCallCount()).To(Equal(1))
				passedPackage, passedExistingResources, passedReader, passedReaderLength := fakeCloudControllerClient.UploadBitsPackageArgsForCall(0)
				Expect(passedPackage).To(Equal(ccv3.Package(appPkg)))
				Expect(passedExistingResources).To(ConsistOf(ccv3.Resource{FilePath: "some-resource"}, ccv3.Resource{FilePath: "another-resource"}))
				Expect(passedReader).To(Equal(reader))
				Expect(passedReaderLength).To(Equal(readerLength))
			})
		})

		When("the upload returns an error", func() {
			var err error

			BeforeEach(func() {
				err = errors.New("some-error")
				fakeCloudControllerClient.UploadBitsPackageReturns(ccv3.Package{}, ccv3.Warnings{"upload-warning-1", "upload-warning-2"}, err)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(err))
				Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2"))
			})
		})
	})

	Describe("PollPackage", func() {
		Context("Polling Behavior", func() {
			var (
				pkg Package

				appPkg     Package
				warnings   Warnings
				executeErr error
			)

			BeforeEach(func() {
				pkg = Package{
					GUID: "some-pkg-guid",
				}

				warnings = nil
				executeErr = nil

				// putting this here so the tests don't hang on polling
				fakeCloudControllerClient.GetPackageReturns(
					ccv3.Package{
						GUID:  "some-pkg-guid",
						State: constant.PackageReady,
					},
					ccv3.Warnings{},
					nil,
				)
			})

			JustBeforeEach(func() {
				appPkg, warnings, executeErr = actor.PollPackage(pkg)
			})

			When("the polling errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("Fake error during polling")
					fakeCloudControllerClient.GetPackageReturns(
						ccv3.Package{},
						ccv3.Warnings{"some-get-pkg-warning"},
						expectedErr,
					)
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("some-get-pkg-warning"))
				})
			})

			When("the polling is successful", func() {
				It("returns the package", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					expectedPackage := ccv3.Package{
						GUID:  "some-pkg-guid",
						State: constant.PackageReady,
					}

					Expect(appPkg).To(Equal(Package(expectedPackage)))
					Expect(fakeCloudControllerClient.GetPackageCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetPackageArgsForCall(0)).To(Equal("some-pkg-guid"))
				})
			})
		})

		DescribeTable("Polling states",
			func(finalState constant.PackageState, expectedErr error) {
				fakeCloudControllerClient.GetPackageReturns(
					ccv3.Package{GUID: "some-pkg-guid", State: constant.PackageAwaitingUpload},
					ccv3.Warnings{"poll-package-warning"},
					nil,
				)

				fakeCloudControllerClient.GetPackageReturnsOnCall(
					1,
					ccv3.Package{State: finalState},
					ccv3.Warnings{"poll-package-warning"},
					nil,
				)

				_, tableWarnings, err := actor.PollPackage(Package{
					GUID: "some-pkg-guid",
				})

				if expectedErr == nil {
					Expect(err).ToNot(HaveOccurred())
				} else {
					Expect(err).To(MatchError(expectedErr))
				}

				Expect(tableWarnings).To(ConsistOf("poll-package-warning", "poll-package-warning"))

				Expect(fakeCloudControllerClient.GetPackageCallCount()).To(Equal(2))
				Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(2))
			},

			Entry("READY", constant.PackageReady, nil),
			Entry("FAILED", constant.PackageFailed, actionerror.PackageProcessingFailedError{}),
			Entry("EXPIRED", constant.PackageExpired, actionerror.PackageProcessingExpiredError{}),
		)
	})
})
