package v7action_test

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"

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
		actor = NewActor(fakeCloudControllerClient, fakeConfig, fakeSharedActor, nil, nil, nil)
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
				Expect(packages).To(Equal([]resources.Package{
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
					ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.CreatedAtDescendingOrder}},
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

	Describe("GetNewestReadyPackageForApplication", func() {
		var (
			app        resources.Application
			executeErr error

			warnings Warnings
		)

		BeforeEach(func() {
			app = resources.Application{
				GUID: "some-app-guid",
				Name: "some-app",
			}
		})

		When("the GetNewestReadyPackageForApplication call succeeds", func() {
			When("the cloud controller finds a package", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetPackagesReturns(
						[]resources.Package{
							{
								GUID:      "some-package-guid-1",
								State:     constant.PackageReady,
								CreatedAt: "2017-08-14T21:16:42Z",
							},
							{
								GUID:      "some-package-guid-2",
								State:     constant.PackageReady,
								CreatedAt: "2017-08-16T00:18:24Z",
							},
						},
						ccv3.Warnings{"get-application-packages-warning"},
						nil,
					)
				})
				It("gets the most recent package for the given app guid that has a ready state", func() {
					expectedPackage, warnings, err := actor.GetNewestReadyPackageForApplication(app)

					Expect(fakeCloudControllerClient.GetPackagesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetPackagesArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{"some-app-guid"}},
						ccv3.Query{Key: ccv3.StatesFilter, Values: []string{"READY"}},
						ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.CreatedAtDescendingOrder}},
					))

					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-application-packages-warning"))
					Expect(expectedPackage).To(Equal(resources.Package{
						GUID:      "some-package-guid-1",
						State:     constant.PackageReady,
						CreatedAt: "2017-08-14T21:16:42Z",
					},
					))
				})
			})

			When("the cloud controller does not find any packages", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetPackagesReturns(
						[]resources.Package{},
						ccv3.Warnings{"get-application-packages-warning"},
						nil,
					)
				})

				JustBeforeEach(func() {
					_, warnings, executeErr = actor.GetNewestReadyPackageForApplication(app)
				})

				It("returns an error and warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.NoEligiblePackagesError{AppName: "some-app"}))
					Expect(warnings).To(ConsistOf("get-application-packages-warning"))
				})
			})

		})
		When("getting the application packages fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get-packages-error")

				fakeCloudControllerClient.GetPackagesReturns(
					[]resources.Package{},
					ccv3.Warnings{"get-packages-warning"},
					expectedErr,
				)
			})

			It("returns the error", func() {
				_, warnings, err := actor.GetNewestReadyPackageForApplication(app)

				Expect(err).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-packages-warning"))
			})
		})
	})

	Describe("CreateDockerPackageByApplicationNameAndSpace", func() {
		var (
			dockerPackage resources.Package
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
					Expect(dockerPackage).To(Equal(resources.Package(expectedPackage)))

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
			pkg        resources.Package
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			bitsPath = ""
			pkg = resources.Package{}
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
										Expect(pkg).To(Equal(resources.Package(expectedPackage)))

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

											// hacky, get packages is called an extra time cause the
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

			When("bits path is an archive", func() {
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

					err = zipit(tempFilePath, bitsPath, "")
					Expect(err).NotTo(HaveOccurred())
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

			When("bits path is symlink to an archive", func() {
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

					err = zipit(tempArchiveFilePath, archivePath, "")
					Expect(err).NotTo(HaveOccurred())
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

	Describe("CreateBitsPackageByApplication", func() {
		var (
			appGUID string

			pkg        resources.Package
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
			pkg              resources.Package
			matchedResources []sharedaction.V3Resource
			reader           io.Reader
			readerLength     int64

			appPkg     resources.Package
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			pkg = resources.Package{GUID: "some-package-guid"}

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
				fakeCloudControllerClient.UploadBitsPackageReturns(resources.Package{GUID: "some-package-guid"}, ccv3.Warnings{"upload-warning-1", "upload-warning-2"}, nil)
			})

			It("returns all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2"))
				Expect(appPkg).To(Equal(resources.Package{GUID: "some-package-guid"}))

				Expect(fakeCloudControllerClient.UploadBitsPackageCallCount()).To(Equal(1))
				passedPackage, passedExistingResources, passedReader, passedReaderLength := fakeCloudControllerClient.UploadBitsPackageArgsForCall(0)
				Expect(passedPackage).To(Equal(resources.Package(appPkg)))
				Expect(passedExistingResources).To(ConsistOf(ccv3.Resource{FilePath: "some-resource"}, ccv3.Resource{FilePath: "another-resource"}))
				Expect(passedReader).To(Equal(reader))
				Expect(passedReaderLength).To(Equal(readerLength))
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
				pkg resources.Package

				appPkg     resources.Package
				warnings   Warnings
				executeErr error
			)

			BeforeEach(func() {
				pkg = resources.Package{
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

					Expect(appPkg).To(Equal(resources.Package(expectedPackage)))
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

				_, tableWarnings, err := actor.PollPackage(resources.Package{
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

	Describe("CopyPackage", func() {
		var (
			sourceApp  resources.Application
			targetApp  resources.Application
			pkg        resources.Package
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			targetApp = resources.Application{
				GUID: "target-app-guid",
				Name: "target-app",
			}

			fakeCloudControllerClient.GetPackagesReturns(
				[]resources.Package{{GUID: "source-package-guid"}},
				ccv3.Warnings{"get-source-package-warning"},
				nil,
			)

			fakeCloudControllerClient.CopyPackageReturns(
				resources.Package{GUID: "target-package-guid"},
				ccv3.Warnings{"copy-package-warning"},
				nil,
			)

			fakeCloudControllerClient.GetPackageReturnsOnCall(0,
				resources.Package{State: constant.PackageCopying, GUID: "target-package-guid"},
				ccv3.Warnings{"get-package-warning-copying"},
				nil,
			)
			fakeCloudControllerClient.GetPackageReturnsOnCall(1,
				resources.Package{State: constant.PackageReady, GUID: "target-package-guid"},
				ccv3.Warnings{"get-package-warning-ready"},
				nil,
			)
		})

		JustBeforeEach(func() {
			pkg, warnings, executeErr = actor.CopyPackage(sourceApp, targetApp)
		})

		When("getting the source package fails", func() {
			var err error

			BeforeEach(func() {
				err = errors.New("get-package-error")
				fakeCloudControllerClient.GetPackagesReturns(
					[]resources.Package{},
					ccv3.Warnings{"get-source-package-warning"},
					err,
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(err))
				Expect(warnings).To(ConsistOf("get-source-package-warning"))

				queries := fakeCloudControllerClient.GetPackagesArgsForCall(0)
				Expect(queries).To(Equal([]ccv3.Query{
					ccv3.Query{
						Key:    ccv3.AppGUIDFilter,
						Values: []string{sourceApp.GUID},
					},
					ccv3.Query{
						Key:    ccv3.StatesFilter,
						Values: []string{string(constant.PackageReady)},
					},
					ccv3.Query{
						Key:    ccv3.OrderBy,
						Values: []string{ccv3.CreatedAtDescendingOrder},
					},
				}))
			})
		})

		When("copying the package fails", func() {
			var err error

			BeforeEach(func() {
				err = errors.New("copy-package-error")
				fakeCloudControllerClient.CopyPackageReturns(
					resources.Package{GUID: "target-package-guid"},
					ccv3.Warnings{"copy-package-warning"},
					err,
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(err))
				Expect(warnings).To(ConsistOf("get-source-package-warning", "copy-package-warning"))

				sourcePkgGUID, appGUID := fakeCloudControllerClient.CopyPackageArgsForCall(0)
				Expect(sourcePkgGUID).To(Equal("source-package-guid"))
				Expect(appGUID).To(Equal(targetApp.GUID))
			})
		})

		It("polls to make sure the package has finished copying", func() {
			Expect(executeErr).To(Not(HaveOccurred()))
			Expect(fakeCloudControllerClient.GetPackageCallCount()).To(Equal(2))
		})

		When("the package fails to copy while polling the package", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetPackageReturnsOnCall(0,
					resources.Package{State: constant.PackageFailed},
					ccv3.Warnings{"get-package-warning-copying"},
					nil,
				)

			})
			It("fails", func() {
				Expect(fakeCloudControllerClient.GetPackageCallCount()).To(Equal(1))
				Expect(executeErr).To(MatchError(actionerror.PackageProcessingFailedError{}))
			})
		})

		It("returns all warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(warnings).To(ConsistOf("get-source-package-warning", "copy-package-warning", "get-package-warning-copying", "get-package-warning-ready"))
			Expect(pkg).To(Equal(resources.Package{State: constant.PackageReady, GUID: "target-package-guid"}))
		})
	})
})
