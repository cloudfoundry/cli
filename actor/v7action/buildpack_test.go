package v7action_test

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Buildpack", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _, _ = NewTestActor()
	})

	Describe("GetBuildpackByNameAndStack", func() {
		var (
			buildpackName  = "buildpack-1"
			buildpackStack = "stack-name"
			buildpack      resources.Buildpack
			warnings       Warnings
			executeErr     error
		)

		JustBeforeEach(func() {
			buildpack, warnings, executeErr = actor.GetBuildpackByNameAndStack(buildpackName, buildpackStack)
		})

		When("getting buildpacks fails", func() {
			BeforeEach(func() {

				buildpackStack = "real-good-stack"
				fakeCloudControllerClient.GetBuildpacksReturns(
					nil,
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("some-error"))
			})

			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
				Expect(fakeCloudControllerClient.GetBuildpacksCallCount()).To(Equal(1))
				queries := fakeCloudControllerClient.GetBuildpacksArgsForCall(0)
				Expect(queries).To(ConsistOf(
					ccv3.Query{
						Key:    ccv3.NameFilter,
						Values: []string{buildpackName},
					},
					ccv3.Query{
						Key:    ccv3.StackFilter,
						Values: []string{buildpackStack},
					},
				))
			})
		})

		When("multiple buildpacks with stacks are returned", func() {
			BeforeEach(func() {
				ccBuildpacks := []resources.Buildpack{
					{Name: buildpackName, Stack: "a-real-stack", Position: types.NullInt{Value: 1, IsSet: true}},
					{Name: buildpackName, Stack: "another-stack", Position: types.NullInt{Value: 2, IsSet: true}},
				}

				fakeCloudControllerClient.GetBuildpacksReturns(
					ccBuildpacks,
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					nil)
			})

			It("returns warnings and MultipleBuildpacksFoundError", func() {
				Expect(executeErr).To(MatchError(actionerror.MultipleBuildpacksFoundError{BuildpackName: buildpackName}))
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
			})

		})

		When("multiple buildpacks including one with no stack are returned", func() {
			BeforeEach(func() {
				ccBuildpacks := []resources.Buildpack{
					{GUID: "buildpack-1-guid", Name: "buildpack-1", Stack: "a-real-stack", Position: types.NullInt{Value: 1, IsSet: true}},
					{GUID: "buildpack-2-guid", Name: "buildpack-2", Stack: "", Position: types.NullInt{Value: 2, IsSet: true}},
				}

				fakeCloudControllerClient.GetBuildpacksReturns(
					ccBuildpacks,
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					nil)
			})

			It("returns the nil stack buildpack and any warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(buildpack).To(Equal(resources.Buildpack{Name: "buildpack-2", GUID: "buildpack-2-guid", Stack: "", Position: types.NullInt{Value: 2, IsSet: true}}))
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
			})
		})

		When("zero buildpacks are returned", func() {
			BeforeEach(func() {
				var ccBuildpacks []resources.Buildpack

				fakeCloudControllerClient.GetBuildpacksReturns(
					ccBuildpacks,
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					nil)
			})

			It("returns warnings and a BuildpackNotFoundError", func() {
				Expect(executeErr).To(MatchError(actionerror.BuildpackNotFoundError{BuildpackName: buildpackName, StackName: buildpackStack}))
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
			})
		})

		When("getting buildpacks is successful", func() {
			When("No stack is specified", func() {
				BeforeEach(func() {
					buildpackStack = ""
					buildpackName = "my-buildpack"

					ccBuildpack := resources.Buildpack{Name: "my-buildpack", GUID: "some-guid"}
					fakeCloudControllerClient.GetBuildpacksReturns(
						[]resources.Buildpack{ccBuildpack},
						ccv3.Warnings{"some-warning-1", "some-warning-2"},
						nil)
				})

				It("Returns the proper buildpack", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
					Expect(buildpack).To(Equal(resources.Buildpack{Name: "my-buildpack", GUID: "some-guid"}))
				})

				It("Does not pass a stack query to the client", func() {
					Expect(fakeCloudControllerClient.GetBuildpacksCallCount()).To(Equal(1))
					queries := fakeCloudControllerClient.GetBuildpacksArgsForCall(0)
					Expect(queries).To(ConsistOf(
						ccv3.Query{
							Key:    ccv3.NameFilter,
							Values: []string{buildpackName},
						},
					))
				})
			})

			When("A stack is specified", func() {
				BeforeEach(func() {
					buildpackStack = "good-stack"
					buildpackName = "my-buildpack"

					ccBuildpack := resources.Buildpack{Name: "my-buildpack", GUID: "some-guid", Stack: "good-stack"}
					fakeCloudControllerClient.GetBuildpacksReturns(
						[]resources.Buildpack{ccBuildpack},
						ccv3.Warnings{"some-warning-1", "some-warning-2"},
						nil)
				})

				It("Returns the proper buildpack", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
					Expect(buildpack).To(Equal(resources.Buildpack{Name: "my-buildpack", GUID: "some-guid", Stack: "good-stack"}))
				})

				It("Does pass a stack query to the client", func() {
					Expect(fakeCloudControllerClient.GetBuildpacksCallCount()).To(Equal(1))
					queries := fakeCloudControllerClient.GetBuildpacksArgsForCall(0)
					Expect(queries).To(ConsistOf(
						ccv3.Query{
							Key:    ccv3.NameFilter,
							Values: []string{buildpackName},
						},
						ccv3.Query{
							Key:    ccv3.StackFilter,
							Values: []string{buildpackStack},
						},
					))
				})
			})
		})
	})

	Describe("GetBuildpacks", func() {
		var (
			buildpacks    []resources.Buildpack
			warnings      Warnings
			executeErr    error
			labelSelector string
		)

		JustBeforeEach(func() {
			buildpacks, warnings, executeErr = actor.GetBuildpacks(labelSelector)
		})

		It("calls CloudControllerClient.GetBuildpacks()", func() {
			Expect(fakeCloudControllerClient.GetBuildpacksCallCount()).To(Equal(1))
		})

		When("a label selector is not provided", func() {
			BeforeEach(func() {
				labelSelector = ""
			})
			It("only passes through a OrderBy query to the CloudControllerClient", func() {
				positionQuery := ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.PositionOrder}}
				Expect(fakeCloudControllerClient.GetBuildpacksArgsForCall(0)).To(ConsistOf(positionQuery))
			})
		})

		When("a label selector is provided", func() {
			BeforeEach(func() {
				labelSelector = "some-label-selector"
			})

			It("passes the labels selector through", func() {
				labelQuery := ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{labelSelector}}
				Expect(fakeCloudControllerClient.GetBuildpacksArgsForCall(0)).To(ContainElement(labelQuery))
			})
		})

		When("getting buildpacks fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetBuildpacksReturns(
					nil,
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("some-error"))
			})

			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
			})
		})

		When("getting buildpacks is successful", func() {
			BeforeEach(func() {
				ccBuildpacks := []resources.Buildpack{
					{Name: "buildpack-1", Position: types.NullInt{Value: 1, IsSet: true}},
					{Name: "buildpack-2", Position: types.NullInt{Value: 2, IsSet: true}},
				}

				fakeCloudControllerClient.GetBuildpacksReturns(
					ccBuildpacks,
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					nil)
			})

			It("returns the buildpacks and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
				Expect(buildpacks).To(Equal([]resources.Buildpack{
					{Name: "buildpack-1", Position: types.NullInt{Value: 1, IsSet: true}},
					{Name: "buildpack-2", Position: types.NullInt{Value: 2, IsSet: true}},
				}))
			})
		})
	})

	Describe("CreateBuildpack", func() {
		var (
			buildpack  resources.Buildpack
			warnings   Warnings
			executeErr error
			bp         resources.Buildpack
		)

		JustBeforeEach(func() {
			buildpack, warnings, executeErr = actor.CreateBuildpack(bp)
		})

		When("creating a buildpack fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateBuildpackReturns(
					resources.Buildpack{},
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("some-error"))
			})

			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
				Expect(buildpack).To(Equal(resources.Buildpack{}))
			})
		})

		When("creating a buildpack is successful", func() {
			var returnBuildpack resources.Buildpack
			BeforeEach(func() {
				bp = resources.Buildpack{Name: "some-name", Stack: "some-stack"}
				returnBuildpack = resources.Buildpack{GUID: "some-guid", Name: "some-name", Stack: "some-stack"}
				fakeCloudControllerClient.CreateBuildpackReturns(
					resources.Buildpack(returnBuildpack),
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					nil)
			})

			It("returns the buildpacks and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
				Expect(buildpack).To(Equal(returnBuildpack))

				Expect(fakeCloudControllerClient.CreateBuildpackCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateBuildpackArgsForCall(0)).To(Equal(resources.Buildpack(bp)))
			})
		})
	})

	Describe("UpdateBuildpackByNameAndStack", func() {
		var (
			buildpackName  = "my-buildpack"
			buildpackStack = "my-stack"
			buildpack      = resources.Buildpack{
				Stack: "new-stack",
			}

			retBuildpack resources.Buildpack
			warnings     Warnings
			executeErr   error
		)

		JustBeforeEach(func() {
			retBuildpack, warnings, executeErr = actor.UpdateBuildpackByNameAndStack(buildpackName, buildpackStack, buildpack)
		})

		When("it is successful", func() {
			var updatedBuildpack resources.Buildpack
			BeforeEach(func() {
				foundBuildpack := resources.Buildpack{GUID: "a guid", Stack: ""}
				updatedBuildpack = resources.Buildpack{GUID: "a guid", Stack: "new-stack"}
				fakeCloudControllerClient.GetBuildpacksReturns([]resources.Buildpack{foundBuildpack}, ccv3.Warnings{"warning-1"}, nil)
				fakeCloudControllerClient.UpdateBuildpackReturns(resources.Buildpack(updatedBuildpack), ccv3.Warnings{"warning-2"}, nil)
			})

			It("returns the updated buildpack and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(retBuildpack).To(Equal(updatedBuildpack))

				queries := fakeCloudControllerClient.GetBuildpacksArgsForCall(0)
				Expect(queries).To(ConsistOf(
					ccv3.Query{
						Key:    ccv3.NameFilter,
						Values: []string{buildpackName},
					},
					ccv3.Query{
						Key:    ccv3.StackFilter,
						Values: []string{buildpackStack},
					},
				))

				paramBuildpack := fakeCloudControllerClient.UpdateBuildpackArgsForCall(0)
				Expect(paramBuildpack).To(Equal(resources.Buildpack{
					GUID:  "a guid",
					Stack: "new-stack",
				}))
			})
		})

		When("The get fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetBuildpacksReturns([]resources.Buildpack{}, ccv3.Warnings{"warning-1"}, errors.New("whoops"))
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1"))
				Expect(retBuildpack).To(Equal(resources.Buildpack{}))
			})
		})

		When("The update fails", func() {
			BeforeEach(func() {
				ccBuildpack := resources.Buildpack{GUID: "a guid", Stack: "old-stack"}
				fakeCloudControllerClient.GetBuildpacksReturns([]resources.Buildpack{ccBuildpack}, ccv3.Warnings{"warning-1"}, nil)
				fakeCloudControllerClient.UpdateBuildpackReturns(resources.Buildpack{}, ccv3.Warnings{"warning-2"}, errors.New("whoops"))
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(retBuildpack).To(Equal(resources.Buildpack{}))
			})
		})

	})

	Describe("UploadBuildpack", func() {
		var (
			bpFile     io.Reader
			bpFilePath string
			fakePb     *v7actionfakes.FakeSimpleProgressBar

			jobURL     ccv3.JobURL
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			bpFile = strings.NewReader("")
			fakePb = new(v7actionfakes.FakeSimpleProgressBar)
			fakePb.InitializeReturns(bpFile, 66, nil)
		})

		JustBeforeEach(func() {
			bpFilePath = "tmp/buildpack.zip"
			jobURL, warnings, executeErr = actor.UploadBuildpack("some-bp-guid", bpFilePath, fakePb)
		})

		It("tracks the progress of the upload", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakePb.InitializeCallCount()).To(Equal(1))
			Expect(fakePb.InitializeArgsForCall(0)).To(Equal(bpFilePath))
			Expect(fakePb.TerminateCallCount()).To(Equal(1))
		})

		When("reading the file errors", func() {
			BeforeEach(func() {
				fakePb.InitializeReturns(bpFile, 66, os.ErrNotExist)
			})

			It("returns the err", func() {
				Expect(executeErr).To(Equal(os.ErrNotExist))
			})
		})

		When("the upload errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadBuildpackReturns(
					ccv3.JobURL(""),
					ccv3.Warnings{"some-upload-warning"},
					errors.New("some-upload-error"),
				)
			})

			It("returns warnings and errors", func() {
				Expect(warnings).To(ConsistOf("some-upload-warning"))
				Expect(executeErr).To(MatchError("some-upload-error"))
			})
		})

		When("the cc returns an error because the buildpack and stack combo already exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadBuildpackReturns(
					ccv3.JobURL(""),
					ccv3.Warnings{"some-upload-warning"},
					ccerror.BuildpackAlreadyExistsForStackError{Message: "ya blew it"},
				)
			})

			It("returns warnings and a BuildpackAlreadyExistsForStackError", func() {
				Expect(warnings).To(ConsistOf("some-upload-warning"))
				Expect(executeErr).To(MatchError(actionerror.BuildpackAlreadyExistsForStackError{Message: "ya blew it"}))
			})
		})

		When("the upload is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadBuildpackReturns(
					ccv3.JobURL("http://example.com/some-job-url"),
					ccv3.Warnings{"some-create-warning"},
					nil,
				)
			})

			It("uploads the buildpack and returns the Job URL and any warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.UploadBuildpackCallCount()).To(Equal(1))
				guid, path, pbReader, size := fakeCloudControllerClient.UploadBuildpackArgsForCall(0)
				Expect(guid).To(Equal("some-bp-guid"))
				Expect(size).To(Equal(int64(66)))
				Expect(path).To(Equal(bpFilePath))
				Expect(pbReader).To(Equal(bpFile))
				Expect(jobURL).To(Equal(ccv3.JobURL("http://example.com/some-job-url")))
				Expect(warnings).To(ConsistOf("some-create-warning"))
			})
		})
	})

	Describe("PrepareBuildpackBits", func() {
		var (
			inPath         string
			outPath        string
			tmpDirPath     string
			fakeDownloader *v7actionfakes.FakeDownloader

			executeErr error
		)

		BeforeEach(func() {
			fakeDownloader = new(v7actionfakes.FakeDownloader)
		})

		JustBeforeEach(func() {
			outPath, executeErr = actor.PrepareBuildpackBits(inPath, tmpDirPath, fakeDownloader)
		})

		When("the buildpack path is a url", func() {
			BeforeEach(func() {
				inPath = "http://buildpacks.com/a.zip"
				fakeDownloader = new(v7actionfakes.FakeDownloader)

				var err error
				tmpDirPath, err = ioutil.TempDir("", "buildpackdir-")
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				Expect(os.RemoveAll(tmpDirPath)).ToNot(HaveOccurred())
			})

			When("downloading the file succeeds", func() {
				BeforeEach(func() {
					fakeDownloader.DownloadReturns("/tmp/buildpackdir-100/a.zip", nil)
				})

				It("downloads the buildpack to a local file", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeDownloader.DownloadCallCount()).To(Equal(1))

					inputPath, inputTmpDirPath := fakeDownloader.DownloadArgsForCall(0)
					Expect(inputPath).To(Equal("http://buildpacks.com/a.zip"))
					Expect(inputTmpDirPath).To(Equal(tmpDirPath))
				})
			})

			When("downloading the file fails", func() {
				BeforeEach(func() {
					fakeDownloader.DownloadReturns("", errors.New("some-download-error"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("some-download-error"))
				})
			})
		})

		When("the buildpack path points to a directory", func() {
			var tempFile *os.File
			BeforeEach(func() {
				var err error
				inPath, err = ioutil.TempDir("", "buildpackdir-")
				Expect(err).ToNot(HaveOccurred())

				tempFile, err = ioutil.TempFile(inPath, "foo")
				Expect(err).ToNot(HaveOccurred())

				tmpDirPath, err = ioutil.TempDir("", "buildpackdir-")
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				tempFile.Close()
				Expect(os.RemoveAll(inPath)).ToNot(HaveOccurred())
				Expect(os.RemoveAll(tmpDirPath)).ToNot(HaveOccurred())
			})

			It("returns a path to the zipped directory", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeDownloader.DownloadCallCount()).To(Equal(0))

				Expect(filepath.Base(outPath)).To(Equal(filepath.Base(inPath) + ".zip"))
			})
		})

		When("the buildpack path points to an empty directory", func() {
			BeforeEach(func() {
				var err error
				inPath, err = ioutil.TempDir("", "some-empty-dir")
				Expect(err).ToNot(HaveOccurred())

				tmpDirPath, err = ioutil.TempDir("", "buildpackdir-")
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				Expect(os.RemoveAll(inPath)).ToNot(HaveOccurred())
				Expect(os.RemoveAll(tmpDirPath)).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.EmptyBuildpackDirectoryError{Path: inPath}))
			})
		})

		When("the buildpack path points to a zip file", func() {
			BeforeEach(func() {
				inPath = "/foo/buildpacks/a.zip"
			})

			It("returns the local filepath", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeDownloader.DownloadCallCount()).To(Equal(0))
				Expect(outPath).To(Equal("/foo/buildpacks/a.zip"))
			})
		})
	})

	Describe("Zipit", func() {
		//tested in buildpack_linux_test.go and buildpack_windows_test.go
		var (
			source string
			target string

			executeErr error
		)

		JustBeforeEach(func() {
			executeErr = Zipit(source, target, "testzip-")
		})

		When("the source directory does not exist", func() {
			BeforeEach(func() {
				source = ""
				target = ""
			})

			It("returns an error", func() {
				Expect(os.IsNotExist(executeErr)).To(BeTrue())
			})
		})
	})

	Describe("DeleteBuildpackByNameAndStack", func() {
		var (
			buildpackName  = "buildpack-name"
			buildpackStack = "buildpack-stack"
			buildpackGUID  = "buildpack-guid"
			jobURL         = "buildpack-delete-job-url"
			warnings       Warnings
			executeErr     error
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.DeleteBuildpackByNameAndStack(buildpackName, buildpackStack)
		})

		When("getting the buildpack fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetBuildpacksReturns(
					[]resources.Buildpack{},
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("api-get-error"))
			})
			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("api-get-error"))
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
				Expect(fakeCloudControllerClient.GetBuildpacksCallCount()).To(Equal(1))
				queries := fakeCloudControllerClient.GetBuildpacksArgsForCall(0)
				Expect(queries).To(ConsistOf(
					ccv3.Query{
						Key:    ccv3.NameFilter,
						Values: []string{buildpackName},
					},
					ccv3.Query{
						Key:    ccv3.StackFilter,
						Values: []string{buildpackStack},
					},
				))
			})
		})

		When("getting the buildpack succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetBuildpacksReturns(
					[]resources.Buildpack{{GUID: buildpackGUID, Name: buildpackName, Stack: buildpackStack}},
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					nil)
			})
			When("deleting a buildpack fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteBuildpackReturns(
						"",
						ccv3.Warnings{"some-warning-3", "some-warning-4"},
						errors.New("api-delete-error"))
				})

				It("returns warnings and error", func() {
					Expect(executeErr).To(MatchError("api-delete-error"))
					Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2", "some-warning-3", "some-warning-4"))
					Expect(fakeCloudControllerClient.DeleteBuildpackCallCount()).To(Equal(1))
					paramGUID := fakeCloudControllerClient.DeleteBuildpackArgsForCall(0)
					Expect(paramGUID).To(Equal(buildpackGUID))
				})
			})

			When("deleting the buildpack is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteBuildpackReturns(
						ccv3.JobURL(jobURL),
						ccv3.Warnings{"some-warning-3", "some-warning-4"},
						nil)
				})

				When("polling the job fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.PollJobReturns(
							ccv3.Warnings{"some-warning-5", "some-warning-6"},
							errors.New("api-poll-job-error"))
					})
					It("returns warnings and an error", func() {
						Expect(executeErr).To(MatchError("api-poll-job-error"))
						Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2", "some-warning-3", "some-warning-4", "some-warning-5", "some-warning-6"))
						Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
						paramURL := fakeCloudControllerClient.PollJobArgsForCall(0)
						Expect(paramURL).To(Equal(ccv3.JobURL(jobURL)))
					})
				})

				When("polling the job succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.PollJobReturns(
							ccv3.Warnings{"some-warning-5", "some-warning-6"},
							nil)
					})
					It("returns all warnings and no error", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2", "some-warning-3", "some-warning-4", "some-warning-5", "some-warning-6"))
						Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
						paramURL := fakeCloudControllerClient.PollJobArgsForCall(0)
						Expect(paramURL).To(Equal(ccv3.JobURL(jobURL)))

					})
				})
			})
		})

	})
})
