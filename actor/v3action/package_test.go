package v3action_test

import (
	"errors"
	"io"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Package Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
		fakeSharedActor           *v3actionfakes.FakeSharedActor
		fakeConfig                *v3actionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v3actionfakes.FakeConfig)
		fakeSharedActor = new(v3actionfakes.FakeSharedActor)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, fakeSharedActor, nil)
	})

	Describe("GetApplicationPackages", func() {
		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.GetPackagesReturns(
					[]resources.Package{
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
					[]resources.Application{},
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
					[]resources.Application{
						{GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-applications-warning"},
					nil,
				)

				fakeCloudControllerClient.GetPackagesReturns(
					[]resources.Package{},
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
					[]resources.Application{},
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
					[]resources.Application{
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
						resources.Package{},
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
					createdPackage := resources.Package{
						DockerImage:    "some-docker-image",
						DockerUsername: "some-username",
						DockerPassword: "some-password",
						GUID:           "some-pkg-guid",
						State:          constant.PackageReady,
						Relationships: resources.Relationships{
							constant.RelationshipTypeApplication: resources.Relationship{
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

					expectedPackage := resources.Package{
						DockerImage:    "some-docker-image",
						DockerUsername: "some-username",
						DockerPassword: "some-password",
						GUID:           "some-pkg-guid",
						State:          constant.PackageReady,
						Relationships: resources.Relationships{
							constant.RelationshipTypeApplication: resources.Relationship{
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
					Expect(fakeCloudControllerClient.CreatePackageArgsForCall(0)).To(Equal(resources.Package{
						Type:           constant.PackageTypeDocker,
						DockerImage:    "some-docker-image",
						DockerUsername: "some-username",
						DockerPassword: "some-password",
						Relationships: resources.Relationships{
							constant.RelationshipTypeApplication: resources.Relationship{GUID: "some-app-guid"},
						},
					}))
				})
			})
		})
	})

	Describe("CreateAndUploadBitsPackageByApplicationNameAndSpace", func() {
		var (
			bitsPath   string
			pkg        Package
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			bitsPath = ""
			pkg = Package{}
			warnings = nil
			executeErr = nil

			// putting this here so the tests don't hang on polling
			fakeCloudControllerClient.GetPackageReturns(
				resources.Package{GUID: "some-pkg-guid", State: constant.PackageReady},
				ccv3.Warnings{},
				nil,
			)
		})

		JustBeforeEach(func() {
			pkg, warnings, executeErr = actor.CreateAndUploadBitsPackageByApplicationNameAndSpace("some-app-name", "some-space-guid", bitsPath)
		})

		When("retrieving the application errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{},
					ccv3.Warnings{"some-app-warning"},
					errors.New("some-get-error"),
				)
			})

			It("returns the warnings and the error", func() {
				Expect(executeErr).To(MatchError("some-get-error"))
				Expect(warnings).To(ConsistOf("some-app-warning"))
			})
		})

		When("the application can be retrieved", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{
						{
							Name: "some-app-name",
							GUID: "some-app-guid",
						},
					},
					ccv3.Warnings{"some-app-warning"},
					nil,
				)
			})

			When("bits path is a directory", func() {
				BeforeEach(func() {
					var err error
					bitsPath, err = os.MkdirTemp("", "example")
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					if bitsPath != "" {
						err := os.RemoveAll(bitsPath)
						Expect(err).ToNot(HaveOccurred())
					}
				})

				It("calls GatherDirectoryResources and ZipDirectoryResources", func() {
					Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(1))
					Expect(fakeSharedActor.ZipDirectoryResourcesCallCount()).To(Equal(1))
				})

				When("gathering resources fails", func() {
					BeforeEach(func() {
						fakeSharedActor.GatherDirectoryResourcesReturns(nil, errors.New("some-gather-error"))
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError("some-gather-error"))
						Expect(warnings).To(ConsistOf("some-app-warning"))
					})
				})

				When("gathering resources succeeds", func() {
					BeforeEach(func() {
						fakeSharedActor.GatherDirectoryResourcesReturns([]sharedaction.Resource{{Filename: "file-1"}, {Filename: "file-2"}}, nil)
					})

					When("zipping gathered resources fails", func() {
						BeforeEach(func() {
							fakeSharedActor.ZipDirectoryResourcesReturns("", errors.New("some-archive-error"))
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError("some-archive-error"))
							Expect(warnings).To(ConsistOf("some-app-warning"))
						})
					})

					When("zipping gathered resources succeeds", func() {
						BeforeEach(func() {
							fakeSharedActor.ZipDirectoryResourcesReturns("zipped-archive", nil)
						})

						When("creating the package fails", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.CreatePackageReturns(
									resources.Package{},
									ccv3.Warnings{"create-package-warning"},
									errors.New("some-create-error"),
								)
							})

							It("returns the error", func() {
								Expect(executeErr).To(MatchError("some-create-error"))
								Expect(warnings).To(ConsistOf("some-app-warning", "create-package-warning"))
							})
						})

						When("creating the package succeeds", func() {
							var createdPackage resources.Package

							BeforeEach(func() {
								createdPackage = resources.Package{
									GUID:  "some-pkg-guid",
									State: constant.PackageAwaitingUpload,
									Relationships: resources.Relationships{
										constant.RelationshipTypeApplication: resources.Relationship{
											GUID: "some-app-guid",
										},
									},
								}

								fakeCloudControllerClient.CreatePackageReturns(
									createdPackage,
									ccv3.Warnings{"some-package-warning"},
									nil,
								)
							})

							It("uploads the package with the path to the zip", func() {
								Expect(fakeCloudControllerClient.UploadPackageCallCount()).To(Equal(1))
								_, zippedArchive := fakeCloudControllerClient.UploadPackageArgsForCall(0)
								Expect(zippedArchive).To(Equal("zipped-archive"))
							})

							When("uploading fails", func() {
								BeforeEach(func() {
									fakeCloudControllerClient.UploadPackageReturns(
										resources.Package{},
										ccv3.Warnings{"upload-package-warning"},
										errors.New("some-error"),
									)
								})

								It("returns the error", func() {
									Expect(executeErr).To(MatchError("some-error"))
									Expect(warnings).To(ConsistOf("some-app-warning", "some-package-warning", "upload-package-warning"))
								})
							})

							When("uploading succeeds", func() {
								BeforeEach(func() {
									fakeCloudControllerClient.UploadPackageReturns(
										resources.Package{},
										ccv3.Warnings{"upload-package-warning"},
										nil,
									)
								})

								When("the polling errors", func() {
									var expectedErr error

									BeforeEach(func() {
										expectedErr = errors.New("Fake error during polling")
										fakeCloudControllerClient.GetPackageReturns(
											resources.Package{},
											ccv3.Warnings{"some-get-pkg-warning"},
											expectedErr,
										)
									})

									It("returns the error and warnings", func() {
										Expect(executeErr).To(MatchError(expectedErr))
										Expect(warnings).To(ConsistOf("some-app-warning", "some-package-warning", "upload-package-warning", "some-get-pkg-warning"))
									})
								})

								When("the polling is successful", func() {
									It("collects all warnings", func() {
										Expect(executeErr).NotTo(HaveOccurred())
										Expect(warnings).To(ConsistOf("some-app-warning", "some-package-warning", "upload-package-warning"))
									})

									It("successfully resolves the app name", func() {
										Expect(executeErr).ToNot(HaveOccurred())

										Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
										Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
											ccv3.Query{Key: ccv3.NameFilter, Values: []string{"some-app-name"}},
											ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"some-space-guid"}},
										))
									})

									It("successfully creates the Package", func() {
										Expect(executeErr).ToNot(HaveOccurred())

										Expect(fakeCloudControllerClient.CreatePackageCallCount()).To(Equal(1))
										inputPackage := fakeCloudControllerClient.CreatePackageArgsForCall(0)
										Expect(inputPackage).To(Equal(resources.Package{
											Type: constant.PackageTypeBits,
											Relationships: resources.Relationships{
												constant.RelationshipTypeApplication: resources.Relationship{GUID: "some-app-guid"},
											},
										}))
									})

									It("returns the package", func() {
										Expect(executeErr).ToNot(HaveOccurred())

										expectedPackage := resources.Package{
											GUID:  "some-pkg-guid",
											State: constant.PackageReady,
										}
										Expect(pkg).To(Equal(Package(expectedPackage)))

										Expect(fakeCloudControllerClient.GetPackageCallCount()).To(Equal(1))
										Expect(fakeCloudControllerClient.GetPackageArgsForCall(0)).To(Equal("some-pkg-guid"))
									})

									DescribeTable("polls until terminal state is reached",
										func(finalState constant.PackageState, expectedErr error) {
											fakeCloudControllerClient.GetPackageReturns(
												resources.Package{GUID: "some-pkg-guid", State: constant.PackageAwaitingUpload},
												ccv3.Warnings{"poll-package-warning"},
												nil,
											)
											fakeCloudControllerClient.GetPackageReturnsOnCall(
												2,
												resources.Package{State: finalState},
												ccv3.Warnings{"poll-package-warning"},
												nil,
											)

											_, tableWarnings, err := actor.CreateAndUploadBitsPackageByApplicationNameAndSpace("some-app-name", "some-space-guid", bitsPath)

											if expectedErr == nil {
												Expect(err).ToNot(HaveOccurred())
											} else {
												Expect(err).To(MatchError(expectedErr))
											}

											Expect(tableWarnings).To(ConsistOf("some-app-warning", "some-package-warning", "upload-package-warning", "poll-package-warning", "poll-package-warning"))

											// hacky, get packages is called an extry time cause the
											// JustBeforeEach executes everything once as well
											Expect(fakeCloudControllerClient.GetPackageCallCount()).To(Equal(3))
											Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(3))
										},

										Entry("READY", constant.PackageReady, nil),
										Entry("FAILED", constant.PackageFailed, actionerror.PackageProcessingFailedError{}),
										Entry("EXPIRED", constant.PackageExpired, actionerror.PackageProcessingExpiredError{}),
									)
								})
							})
						})
					})
				})
			})

			When("bitsPath is blank", func() {
				var oldCurrentDir, appDir string
				BeforeEach(func() {
					var err error
					oldCurrentDir, err = os.Getwd()
					Expect(err).NotTo(HaveOccurred())

					appDir, err = os.MkdirTemp("", "example")
					Expect(err).ToNot(HaveOccurred())

					Expect(os.Chdir(appDir)).NotTo(HaveOccurred())
					appDir, err = os.Getwd()
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					Expect(os.Chdir(oldCurrentDir)).NotTo(HaveOccurred())
					err := os.RemoveAll(appDir)
					Expect(err).ToNot(HaveOccurred())
				})

				It("uses the current working directory", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(1))
					Expect(fakeSharedActor.GatherDirectoryResourcesArgsForCall(0)).To(Equal(appDir))

					Expect(fakeSharedActor.ZipDirectoryResourcesCallCount()).To(Equal(1))
					pathArg, _ := fakeSharedActor.ZipDirectoryResourcesArgsForCall(0)
					Expect(pathArg).To(Equal(appDir))
				})
			})

			When("bits path is an archive", func() {
				BeforeEach(func() {
					var err error
					tempFile, err := os.CreateTemp("", "bits-zip-test")
					Expect(err).ToNot(HaveOccurred())
					Expect(tempFile.Close()).To(Succeed())
					tempFilePath := tempFile.Name()

					bitsPathFile, err := os.CreateTemp("", "example")
					Expect(err).ToNot(HaveOccurred())
					Expect(bitsPathFile.Close()).To(Succeed())
					bitsPath = bitsPathFile.Name()

					err = zipit(tempFilePath, bitsPath, "")
					Expect(err).ToNot(HaveOccurred())
					Expect(os.Remove(tempFilePath)).To(Succeed())
				})

				AfterEach(func() {
					err := os.RemoveAll(bitsPath)
					Expect(err).ToNot(HaveOccurred())
				})

				It("calls GatherArchiveResources and ZipArchiveResources", func() {
					Expect(fakeSharedActor.GatherArchiveResourcesCallCount()).To(Equal(1))
					Expect(fakeSharedActor.ZipArchiveResourcesCallCount()).To(Equal(1))
				})

				When("gathering archive resources fails", func() {
					BeforeEach(func() {
						fakeSharedActor.GatherArchiveResourcesReturns(nil, errors.New("some-archive-resource-error"))
					})
					It("should return an error", func() {
						Expect(executeErr).To(MatchError("some-archive-resource-error"))
						Expect(warnings).To(ConsistOf("some-app-warning"))
					})

				})

				When("gathering resources succeeds", func() {
					BeforeEach(func() {
						fakeSharedActor.GatherArchiveResourcesReturns([]sharedaction.Resource{{Filename: "file-1"}, {Filename: "file-2"}}, nil)
					})

					When("zipping gathered resources fails", func() {
						BeforeEach(func() {
							fakeSharedActor.ZipArchiveResourcesReturns("", errors.New("some-archive-error"))
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError("some-archive-error"))
							Expect(warnings).To(ConsistOf("some-app-warning"))
						})
					})

					When("zipping gathered resources succeeds", func() {
						BeforeEach(func() {
							fakeSharedActor.ZipArchiveResourcesReturns("zipped-archive", nil)
						})

						It("uploads the package", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(warnings).To(ConsistOf("some-app-warning"))

							Expect(fakeCloudControllerClient.UploadPackageCallCount()).To(Equal(1))
							_, archivePathArg := fakeCloudControllerClient.UploadPackageArgsForCall(0)
							Expect(archivePathArg).To(Equal("zipped-archive"))
						})
					})
				})
			})

			When("bits path is a symlink to a directory", func() {
				var tempDir string

				BeforeEach(func() {
					var err error
					tempDir, err = os.MkdirTemp("", "example")
					Expect(err).ToNot(HaveOccurred())

					tempFile, err := os.CreateTemp("", "example-file-")
					Expect(err).ToNot(HaveOccurred())
					Expect(tempFile.Close()).To(Succeed())

					bitsPath = tempFile.Name()
					Expect(os.Remove(bitsPath)).To(Succeed())
					Expect(os.Symlink(tempDir, bitsPath)).To(Succeed())
				})

				AfterEach(func() {
					Expect(os.RemoveAll(tempDir)).To(Succeed())
					Expect(os.Remove(bitsPath)).To(Succeed())
				})

				It("calls GatherDirectoryResources and returns without an error", func() {
					Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(1))
					Expect(fakeSharedActor.GatherDirectoryResourcesArgsForCall(0)).To(Equal(bitsPath))
					Expect(executeErr).ToNot(HaveOccurred())
				})
			})

			When("bits path is symlink to an archive", func() {
				var archivePath string

				BeforeEach(func() {
					var err error
					tempArchiveFile, err := os.CreateTemp("", "bits-zip-test")
					Expect(err).ToNot(HaveOccurred())
					Expect(tempArchiveFile.Close()).To(Succeed())
					tempArchiveFilePath := tempArchiveFile.Name()

					archivePathFile, err := os.CreateTemp("", "example")
					Expect(err).ToNot(HaveOccurred())
					Expect(archivePathFile.Close()).To(Succeed())
					archivePath = archivePathFile.Name()

					err = zipit(tempArchiveFilePath, archivePath, "")
					Expect(err).ToNot(HaveOccurred())
					Expect(os.Remove(tempArchiveFilePath)).To(Succeed())

					tempFile, err := os.CreateTemp("", "example-file-")
					Expect(err).ToNot(HaveOccurred())
					Expect(tempFile.Close()).To(Succeed())

					bitsPath = tempFile.Name()
					Expect(os.Remove(bitsPath)).To(Succeed())
					Expect(os.Symlink(archivePath, bitsPath)).To(Succeed())
				})

				AfterEach(func() {
					Expect(os.Remove(archivePath)).To(Succeed())
					Expect(os.Remove(bitsPath)).To(Succeed())
				})

				It("calls GatherArchiveResources and returns without an error", func() {
					Expect(fakeSharedActor.GatherArchiveResourcesCallCount()).To(Equal(1))
					Expect(fakeSharedActor.GatherArchiveResourcesArgsForCall(0)).To(Equal(bitsPath))
					Expect(executeErr).ToNot(HaveOccurred())
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
					resources.Package{},
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
			var createdPackage resources.Package

			BeforeEach(func() {
				createdPackage = resources.Package{GUID: "some-pkg-guid"}
				fakeCloudControllerClient.CreatePackageReturns(
					createdPackage,
					ccv3.Warnings{"create-package-warning"},
					nil,
				)
			})

			It("returns all warnings and the package", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.CreatePackageCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreatePackageArgsForCall(0)).To(Equal(resources.Package{
					Type: constant.PackageTypeBits,
					Relationships: resources.Relationships{
						constant.RelationshipTypeApplication: resources.Relationship{GUID: appGUID},
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
			matchedResources []sharedaction.Resource
			reader           io.Reader
			readerLength     int64

			appPkg     Package
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			pkg = Package{GUID: "some-package-guid"}

			matchedResources = []sharedaction.Resource{{Filename: "some-resource"}, {Filename: "another-resource"}}
			someString := "who reads these days"
			reader = strings.NewReader(someString)
			readerLength = int64(len([]byte(someString)))
		})

		JustBeforeEach(func() {
			appPkg, warnings, executeErr = actor.UploadBitsPackage(pkg, matchedResources, reader, readerLength)
		})

		When("the upload is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadBitsPackageReturns(resources.Package{GUID: "some-package-guid"}, ccv3.Warnings{"upload-warning-1", "upload-warning-2"}, nil)
			})

			It("passes a ccv3 Resource to the client", func() {
				passedPackage, passedMatchedResources, passedReader, passedReaderLength := fakeCloudControllerClient.UploadBitsPackageArgsForCall(0)
				Expect(passedPackage).To(Equal(resources.Package(appPkg)))
				Expect(passedMatchedResources).To(ConsistOf(ccv3.Resource{FilePath: "some-resource"}, ccv3.Resource{FilePath: "another-resource"}))
				Expect(passedReader).To(Equal(reader))
				Expect(passedReaderLength).To(Equal(readerLength))
			})

			It("returns all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2"))
				Expect(appPkg).To(Equal(Package{GUID: "some-package-guid"}))

				Expect(fakeCloudControllerClient.UploadBitsPackageCallCount()).To(Equal(1))

			})
		})

		When("the upload returns an error", func() {
			var err error

			BeforeEach(func() {
				err = errors.New("some-error")
				fakeCloudControllerClient.UploadBitsPackageReturns(resources.Package{}, ccv3.Warnings{"upload-warning-1", "upload-warning-2"}, err)
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
					resources.Package{
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
						resources.Package{},
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

					expectedPackage := resources.Package{
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
					resources.Package{GUID: "some-pkg-guid", State: constant.PackageAwaitingUpload},
					ccv3.Warnings{"poll-package-warning"},
					nil,
				)

				fakeCloudControllerClient.GetPackageReturnsOnCall(
					1,
					resources.Package{State: finalState},
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
