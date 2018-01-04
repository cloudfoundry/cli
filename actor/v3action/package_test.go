package v3action_test

import (
	"errors"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
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
		Context("when there are no client errors", func() {
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

		Context("when getting the application fails", func() {
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

		Context("when getting the application packages fails", func() {
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

		Context("when the application can't be retrieved", func() {
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

		Context("when the application can be retrieved", func() {
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

			Context("when creating the package fails", func() {
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

			Context("when creating the package succeeds", func() {
				BeforeEach(func() {
					createdPackage := ccv3.Package{
						DockerImage:    "some-docker-image",
						DockerUsername: "some-username",
						DockerPassword: "some-password",
						GUID:           "some-pkg-guid",
						State:          constant.PackageReady,
						Relationships: ccv3.Relationships{
							ccv3.ApplicationRelationship: ccv3.Relationship{
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
							ccv3.ApplicationRelationship: ccv3.Relationship{
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
							ccv3.ApplicationRelationship: ccv3.Relationship{GUID: "some-app-guid"},
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
				ccv3.Package{GUID: "some-pkg-guid", State: constant.PackageReady},
				ccv3.Warnings{},
				nil,
			)
		})

		JustBeforeEach(func() {
			pkg, warnings, executeErr = actor.CreateAndUploadBitsPackageByApplicationNameAndSpace("some-app-name", "some-space-guid", bitsPath)
		})

		Context("when retrieving the application errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{},
					ccv3.Warnings{"some-app-warning"},
					errors.New("some-get-error"),
				)
			})

			It("returns the warnings and the error", func() {
				Expect(executeErr).To(MatchError("some-get-error"))
				Expect(warnings).To(ConsistOf("some-app-warning"))
			})
		})

		Context("when the application can be retrieved", func() {
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

			Context("when bits path is a directory", func() {
				BeforeEach(func() {
					var err error
					bitsPath, err = ioutil.TempDir("", "example")
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

				Context("when gathering resources fails", func() {
					BeforeEach(func() {
						fakeSharedActor.GatherDirectoryResourcesReturns(nil, errors.New("some-gather-error"))
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError("some-gather-error"))
						Expect(warnings).To(ConsistOf("some-app-warning"))
					})
				})

				Context("when gathering resources succeeds", func() {
					BeforeEach(func() {
						fakeSharedActor.GatherDirectoryResourcesReturns([]sharedaction.Resource{{Filename: "file-1"}, {Filename: "file-2"}}, nil)
					})

					Context("when zipping gathered resources fails", func() {
						BeforeEach(func() {
							fakeSharedActor.ZipDirectoryResourcesReturns("", errors.New("some-archive-error"))
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError("some-archive-error"))
							Expect(warnings).To(ConsistOf("some-app-warning"))
						})
					})

					Context("when zipping gathered resources succeeds", func() {
						BeforeEach(func() {
							fakeSharedActor.ZipDirectoryResourcesReturns("zipped-archive", nil)
						})

						Context("when creating the package fails", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.CreatePackageReturns(
									ccv3.Package{},
									ccv3.Warnings{"create-package-warning"},
									errors.New("some-create-error"),
								)
							})

							It("returns the error", func() {
								Expect(executeErr).To(MatchError("some-create-error"))
								Expect(warnings).To(ConsistOf("some-app-warning", "create-package-warning"))
							})
						})

						Context("when creating the package succeeds", func() {
							var createdPackage ccv3.Package

							BeforeEach(func() {
								createdPackage = ccv3.Package{
									GUID:  "some-pkg-guid",
									State: constant.PackageAwaitingUpload,
									Relationships: ccv3.Relationships{
										ccv3.ApplicationRelationship: ccv3.Relationship{
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

							Context("when uploading fails", func() {
								BeforeEach(func() {
									fakeCloudControllerClient.UploadPackageReturns(
										ccv3.Package{},
										ccv3.Warnings{"upload-package-warning"},
										errors.New("some-error"),
									)
								})

								It("returns the error", func() {
									Expect(executeErr).To(MatchError("some-error"))
									Expect(warnings).To(ConsistOf("some-app-warning", "some-package-warning", "upload-package-warning"))
								})
							})

							Context("when uploading succeeds", func() {
								BeforeEach(func() {
									fakeCloudControllerClient.UploadPackageReturns(
										ccv3.Package{},
										ccv3.Warnings{"upload-package-warning"},
										nil,
									)
								})

								Context("when the polling errors", func() {
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
										Expect(warnings).To(ConsistOf("some-app-warning", "some-package-warning", "upload-package-warning", "some-get-pkg-warning"))
									})
								})

								Context("when the polling is successful", func() {
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
										Expect(inputPackage).To(Equal(ccv3.Package{
											Type: constant.PackageTypeBits,
											Relationships: ccv3.Relationships{
												ccv3.ApplicationRelationship: ccv3.Relationship{GUID: "some-app-guid"},
											},
										}))
									})

									It("returns the package", func() {
										Expect(executeErr).ToNot(HaveOccurred())

										expectedPackage := ccv3.Package{
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
												ccv3.Package{GUID: "some-pkg-guid", State: constant.PackageAwaitingUpload},
												ccv3.Warnings{"poll-package-warning"},
												nil,
											)
											fakeCloudControllerClient.GetPackageReturnsOnCall(
												2,
												ccv3.Package{State: finalState},
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

			Context("when bitsPath is blank", func() {
				var oldCurrentDir, appDir string
				BeforeEach(func() {
					var err error
					oldCurrentDir, err = os.Getwd()
					Expect(err).NotTo(HaveOccurred())

					appDir, err = ioutil.TempDir("", "example")
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

			Context("when bits path is an archive", func() {
				BeforeEach(func() {
					var err error
					tempFile, err := ioutil.TempFile("", "bits-zip-test")
					Expect(err).ToNot(HaveOccurred())
					Expect(tempFile.Close()).To(Succeed())
					tempFilePath := tempFile.Name()

					bitsPathFile, err := ioutil.TempFile("", "example")
					Expect(err).ToNot(HaveOccurred())
					Expect(bitsPathFile.Close()).To(Succeed())
					bitsPath = bitsPathFile.Name()

					zipit(tempFilePath, bitsPath, "")
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

				Context("when gathering archive resources fails", func() {
					BeforeEach(func() {
						fakeSharedActor.GatherArchiveResourcesReturns(nil, errors.New("some-archive-resource-error"))
					})
					It("should return an error", func() {
						Expect(executeErr).To(MatchError("some-archive-resource-error"))
						Expect(warnings).To(ConsistOf("some-app-warning"))
					})

				})

				Context("when gathering resources succeeds", func() {
					BeforeEach(func() {
						fakeSharedActor.GatherArchiveResourcesReturns([]sharedaction.Resource{{Filename: "file-1"}, {Filename: "file-2"}}, nil)
					})

					Context("when zipping gathered resources fails", func() {
						BeforeEach(func() {
							fakeSharedActor.ZipArchiveResourcesReturns("", errors.New("some-archive-error"))
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError("some-archive-error"))
							Expect(warnings).To(ConsistOf("some-app-warning"))
						})
					})

					Context("when zipping gathered resources succeeds", func() {
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

			Context("when bits path is a symlink to a directory", func() {
				var tempDir string

				BeforeEach(func() {
					var err error
					tempDir, err = ioutil.TempDir("", "example")
					Expect(err).ToNot(HaveOccurred())

					tempFile, err := ioutil.TempFile("", "example-file-")
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

			Context("when bits path is symlink to an archive", func() {
				var archivePath string

				BeforeEach(func() {
					var err error
					tempArchiveFile, err := ioutil.TempFile("", "bits-zip-test")
					Expect(err).ToNot(HaveOccurred())
					Expect(tempArchiveFile.Close()).To(Succeed())
					tempArchiveFilePath := tempArchiveFile.Name()

					archivePathFile, err := ioutil.TempFile("", "example")
					Expect(err).ToNot(HaveOccurred())
					Expect(archivePathFile.Close()).To(Succeed())
					archivePath = archivePathFile.Name()

					zipit(tempArchiveFilePath, archivePath, "")
					Expect(os.Remove(tempArchiveFilePath)).To(Succeed())

					tempFile, err := ioutil.TempFile("", "example-file-")
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
})
