package v7pushaction_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func buildEmptyV3Resource(name string) sharedaction.V3Resource {
	return sharedaction.V3Resource{
		FilePath:    name,
		Checksum:    ccv3.Checksum{Value: fmt.Sprintf("checksum-%s", name)},
		SizeInBytes: 0,
	}
}

var _ = Describe("CreateBitsPackageForApplication", func() {
	var (
		actor           *Actor
		fakeV7Actor     *v7pushactionfakes.FakeV7Actor
		fakeSharedActor *v7pushactionfakes.FakeSharedActor

		returnedPushPlan PushPlan
		paramPlan        PushPlan
		fakeProgressBar  *v7pushactionfakes.FakeProgressBar

		warnings   Warnings
		executeErr error

		events []Event
	)

	BeforeEach(func() {
		actor, fakeV7Actor, fakeSharedActor = getTestPushActor()

		fakeProgressBar = new(v7pushactionfakes.FakeProgressBar)

		fakeSharedActor.ReadArchiveReturns(new(v7pushactionfakes.FakeReadCloser), 0, nil)

		paramPlan = PushPlan{
			Application: v7action.Application{
				GUID: "some-app-guid",
			},
			DockerImageCredentialsNeedsUpdate: false,
		}
	})

	JustBeforeEach(func() {
		events = EventFollower(func(eventStream chan<- *PushEvent) {
			returnedPushPlan, warnings, executeErr = actor.CreateBitsPackageForApplication(paramPlan, eventStream, fakeProgressBar)
		})
	})

	Describe("package upload", func() {
		When("resource match errors ", func() {
			BeforeEach(func() {
				paramPlan.AllResources = []sharedaction.V3Resource{
					buildV3Resource("some-filename"),
				}

				fakeV7Actor.ResourceMatchReturns(
					nil,
					v7action.Warnings{"some-resource-match-warning"},
					errors.New("resource-match-error"))
			})

			It("raises the error", func() {
				Expect(executeErr).To(MatchError("resource-match-error"))
				Expect(events).To(ConsistOf(ResourceMatching))
				Expect(warnings).To(ConsistOf("some-resource-match-warning"))
			})
		})

		When("resource match is successful", func() {
			var (
				matches   []sharedaction.V3Resource
				unmatches []sharedaction.V3Resource
			)

			When("all resources are empty", func() {
				BeforeEach(func() {
					emptyFiles := []sharedaction.V3Resource{
						buildEmptyV3Resource("empty-filename-1"),
						buildEmptyV3Resource("empty-filename-2"),
					}

					paramPlan = PushPlan{
						Application: v7action.Application{
							Name: "some-app",
							GUID: "some-app-guid",
						},
						BitsPath:     "/some-bits-path",
						AllResources: emptyFiles,
					}
				})

				It("skips resource matching", func() {
					Expect(fakeV7Actor.ResourceMatchCallCount()).To(Equal(0))
				})
			})

			When("there are unmatched resources", func() {
				BeforeEach(func() {
					matches = []sharedaction.V3Resource{
						buildV3Resource("some-matching-filename"),
					}

					unmatches = []sharedaction.V3Resource{
						buildV3Resource("some-unmatching-filename"),
					}

					paramPlan = PushPlan{
						Application: v7action.Application{
							Name: "some-app",
							GUID: "some-app-guid",
						},
						BitsPath: "/some-bits-path",
						AllResources: append(
							matches,
							unmatches...,
						),
					}

					fakeV7Actor.ResourceMatchReturns(
						matches,
						v7action.Warnings{"some-good-good-resource-match-warnings"},
						nil,
					)
				})

				When("the bits path is an archive", func() {
					BeforeEach(func() {
						paramPlan.Archive = true
					})

					It("creates the archive with the unmatched resources", func() {
						Expect(fakeSharedActor.ZipArchiveResourcesCallCount()).To(Equal(1))
						bitsPath, resources := fakeSharedActor.ZipArchiveResourcesArgsForCall(0)
						Expect(bitsPath).To(Equal("/some-bits-path"))
						Expect(resources).To(HaveLen(1))
						Expect(resources[0].ToV3Resource()).To(Equal(unmatches[0]))
					})
				})

				When("The bits path is a directory", func() {
					It("creates the archive", func() {
						Expect(fakeSharedActor.ZipDirectoryResourcesCallCount()).To(Equal(1))
						bitsPath, resources := fakeSharedActor.ZipDirectoryResourcesArgsForCall(0)
						Expect(bitsPath).To(Equal("/some-bits-path"))
						Expect(resources).To(HaveLen(1))
						Expect(resources[0].ToV3Resource()).To(Equal(unmatches[0]))
					})
				})

				When("the archive creation is successful", func() {
					BeforeEach(func() {
						fakeSharedActor.ZipDirectoryResourcesReturns("/some/archive/path", nil)
						fakeV7Actor.UpdateApplicationReturns(
							v7action.Application{
								Name: "some-app",
								GUID: paramPlan.Application.GUID,
							},
							v7action.Warnings{"some-app-update-warnings"},
							nil)
					})

					It("creates the package", func() {
						Expect(fakeV7Actor.CreateBitsPackageByApplicationCallCount()).To(Equal(1))
						Expect(fakeV7Actor.CreateBitsPackageByApplicationArgsForCall(0)).To(Equal("some-app-guid"))
					})

					When("the package creation is successful", func() {
						BeforeEach(func() {
							fakeV7Actor.CreateBitsPackageByApplicationReturns(v7action.Package{GUID: "some-guid"}, v7action.Warnings{"some-create-package-warning"}, nil)
						})

						It("reads the archive", func() {
							Expect(fakeSharedActor.ReadArchiveCallCount()).To(Equal(1))
							Expect(fakeSharedActor.ReadArchiveArgsForCall(0)).To(Equal("/some/archive/path"))
						})

						When("reading the archive is successful", func() {
							BeforeEach(func() {
								fakeReadCloser := new(v7pushactionfakes.FakeReadCloser)
								fakeSharedActor.ReadArchiveReturns(fakeReadCloser, 6, nil)
							})

							It("uploads the bits package", func() {
								Expect(fakeV7Actor.UploadBitsPackageCallCount()).To(Equal(1))
								pkg, resource, _, size := fakeV7Actor.UploadBitsPackageArgsForCall(0)

								Expect(pkg).To(Equal(v7action.Package{GUID: "some-guid"}))
								Expect(resource).To(Equal(matches))
								Expect(size).To(BeNumerically("==", 6))
							})

							When("the upload is successful", func() {
								BeforeEach(func() {
									fakeV7Actor.UploadBitsPackageReturns(v7action.Package{GUID: "some-guid"}, v7action.Warnings{"some-upload-package-warning"}, nil)
								})

								It("returns an upload complete event and warnings", func() {
									Expect(events).To(ConsistOf(ResourceMatching, CreatingPackage, CreatingArchive, ReadingArchive, UploadingApplicationWithArchive, UploadWithArchiveComplete))
									Expect(warnings).To(ConsistOf("some-good-good-resource-match-warnings", "some-create-package-warning", "some-upload-package-warning"))
								})

								When("the upload errors", func() {
									When("the upload error is a retryable error", func() {
										var someErr error

										BeforeEach(func() {
											someErr = errors.New("I AM A BANANA")
											fakeV7Actor.UploadBitsPackageReturns(v7action.Package{}, v7action.Warnings{"upload-warnings-1", "upload-warnings-2"}, ccerror.PipeSeekError{Err: someErr})
										})

										It("should send a RetryUpload event and retry uploading", func() {
											Expect(events).To(ConsistOf(
												ResourceMatching, CreatingPackage, CreatingArchive,
												ReadingArchive, UploadingApplicationWithArchive, RetryUpload,
												ReadingArchive, UploadingApplicationWithArchive, RetryUpload,
												ReadingArchive, UploadingApplicationWithArchive, RetryUpload,
											))

											Expect(warnings).To(ConsistOf("some-good-good-resource-match-warnings", "some-create-package-warning", "upload-warnings-1", "upload-warnings-2", "upload-warnings-1", "upload-warnings-2", "upload-warnings-1", "upload-warnings-2"))

											Expect(fakeV7Actor.UploadBitsPackageCallCount()).To(Equal(3))
											Expect(executeErr).To(MatchError(actionerror.UploadFailedError{Err: someErr}))
										})

									})

									When("the upload error is not a retryable error", func() {
										BeforeEach(func() {
											fakeV7Actor.UploadBitsPackageReturns(v7action.Package{}, v7action.Warnings{"upload-warnings-1", "upload-warnings-2"}, errors.New("dios mio"))
										})

										It("sends warnings and errors, then stops", func() {
											Expect(events).To(ConsistOf(ResourceMatching, CreatingPackage, CreatingArchive, ReadingArchive, UploadingApplicationWithArchive))
											Expect(warnings).To(ConsistOf("some-good-good-resource-match-warnings", "some-create-package-warning", "upload-warnings-1", "upload-warnings-2"))
											Expect(executeErr).To(MatchError("dios mio"))
										})
									})
								})
							})

							When("reading the archive fails", func() {
								BeforeEach(func() {
									fakeSharedActor.ReadArchiveReturns(nil, 0, errors.New("the bits"))
								})

								It("returns an error", func() {
									Expect(events).To(ConsistOf(ResourceMatching, CreatingPackage, CreatingArchive, ReadingArchive))
									Expect(executeErr).To(MatchError("the bits"))
								})
							})
						})

						When("the package creation errors", func() {
							BeforeEach(func() {
								fakeV7Actor.CreateBitsPackageByApplicationReturns(v7action.Package{}, v7action.Warnings{"package-creation-warning"}, errors.New("the package"))
							})

							It("it returns errors and warnings", func() {
								Expect(events).To(ConsistOf(ResourceMatching, CreatingPackage))

								Expect(warnings).To(ConsistOf("some-good-good-resource-match-warnings", "package-creation-warning"))
								Expect(executeErr).To(MatchError("the package"))
							})
						})
					})

					When("the archive creation errors", func() {
						BeforeEach(func() {
							fakeSharedActor.ZipDirectoryResourcesReturns("", errors.New("oh no"))
						})

						It("returns an error and exits", func() {
							Expect(events).To(ConsistOf(ResourceMatching, CreatingPackage, CreatingArchive))
							Expect(executeErr).To(MatchError("oh no"))
						})
					})
				})
			})

			When("All resources are matched", func() {
				BeforeEach(func() {
					matches = []sharedaction.V3Resource{
						buildV3Resource("some-matching-filename"),
					}

					paramPlan = PushPlan{
						Application: v7action.Application{
							Name: "some-app",
							GUID: "some-app-guid",
						},
						BitsPath:     "/some-bits-path",
						AllResources: matches,
					}

					fakeV7Actor.ResourceMatchReturns(
						matches,
						v7action.Warnings{"some-good-good-resource-match-warnings"},
						nil,
					)
				})

				When("package upload succeeds", func() {
					BeforeEach(func() {
						fakeV7Actor.UploadBitsPackageReturns(v7action.Package{GUID: "some-guid"}, v7action.Warnings{"upload-warning"}, nil)
					})

					It("Uploads the package without a zip", func() {
						Expect(fakeSharedActor.ZipArchiveResourcesCallCount()).To(BeZero())
						Expect(fakeSharedActor.ZipDirectoryResourcesCallCount()).To(BeZero())
						Expect(fakeSharedActor.ReadArchiveCallCount()).To(BeZero())

						Expect(events).To(ConsistOf(ResourceMatching, CreatingPackage, UploadingApplication))
						Expect(fakeV7Actor.UploadBitsPackageCallCount()).To(Equal(1))
						_, actualMatchedResources, actualProgressReader, actualSize := fakeV7Actor.UploadBitsPackageArgsForCall(0)

						Expect(actualMatchedResources).To(Equal(matches))
						Expect(actualProgressReader).To(BeNil())
						Expect(actualSize).To(BeZero())
					})
				})

				When("package upload fails", func() {
					BeforeEach(func() {
						fakeV7Actor.UploadBitsPackageReturns(
							v7action.Package{},
							v7action.Warnings{"upload-warning"},
							errors.New("upload-error"),
						)
					})

					It("returns an error", func() {
						Expect(fakeSharedActor.ZipArchiveResourcesCallCount()).To(BeZero())
						Expect(fakeSharedActor.ZipDirectoryResourcesCallCount()).To(BeZero())
						Expect(fakeSharedActor.ReadArchiveCallCount()).To(BeZero())

						Expect(events).To(ConsistOf(ResourceMatching, CreatingPackage, UploadingApplication))
						Expect(fakeV7Actor.UploadBitsPackageCallCount()).To(Equal(1))
						_, actualMatchedResources, actualProgressReader, actualSize := fakeV7Actor.UploadBitsPackageArgsForCall(0)

						Expect(actualMatchedResources).To(Equal(matches))
						Expect(actualProgressReader).To(BeNil())
						Expect(actualSize).To(BeZero())

						Expect(warnings).To(ConsistOf("some-good-good-resource-match-warnings", "upload-warning"))
						Expect(executeErr).To(MatchError("upload-error"))
					})
				})
			})
		})
	})

	Describe("polling package", func() {
		var (
			matches   []sharedaction.V3Resource
			unmatches []sharedaction.V3Resource
		)

		BeforeEach(func() {
			matches = []sharedaction.V3Resource{
				buildV3Resource("some-matching-filename"),
			}

			unmatches = []sharedaction.V3Resource{
				buildV3Resource("some-unmatching-filename"),
			}

			paramPlan = PushPlan{
				Application: v7action.Application{
					Name: "some-app",
					GUID: "some-app-guid",
				},
				BitsPath: "/some-bits-path",
				AllResources: append(
					matches,
					unmatches...,
				),
			}

			fakeV7Actor.ResourceMatchReturns(
				matches,
				v7action.Warnings{"some-good-good-resource-match-warnings"},
				nil,
			)
		})

		When("the the polling is successful", func() {
			BeforeEach(func() {
				fakeV7Actor.PollPackageReturns(v7action.Package{GUID: "some-package-guid"}, v7action.Warnings{"some-poll-package-warning"}, nil)
			})

			It("returns warnings", func() {
				Expect(events).To(ConsistOf(ResourceMatching, CreatingPackage, CreatingArchive, ReadingArchive, UploadingApplicationWithArchive, UploadWithArchiveComplete))
				Expect(warnings).To(ConsistOf("some-good-good-resource-match-warnings", "some-poll-package-warning"))
			})

			It("sets the package guid on push plan", func() {
				Expect(returnedPushPlan.PackageGUID).To(Equal("some-package-guid"))
			})
		})

		When("the the polling returns an error", func() {
			var someErr error

			BeforeEach(func() {
				someErr = errors.New("I AM A BANANA")
				fakeV7Actor.PollPackageReturns(v7action.Package{}, v7action.Warnings{"some-poll-package-warning"}, someErr)
			})

			It("returns errors and warnings", func() {
				Expect(events).To(ConsistOf(ResourceMatching, CreatingPackage, CreatingArchive, ReadingArchive, UploadingApplicationWithArchive, UploadWithArchiveComplete))
				Expect(executeErr).To(MatchError(someErr))
			})
		})
	})
})
