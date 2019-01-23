package v7action_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var _ = Describe("Buildpack", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _ = NewTestActor()
	})

	Describe("GetBuildpacks", func() {
		var (
			buildpacks []Buildpack
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			buildpacks, warnings, executeErr = actor.GetBuildpacks()
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
				ccBuildpacks := []ccv3.Buildpack{
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
				Expect(buildpacks).To(Equal([]Buildpack{
					{Name: "buildpack-1", Position: types.NullInt{Value: 1, IsSet: true}},
					{Name: "buildpack-2", Position: types.NullInt{Value: 2, IsSet: true}},
				}))

				Expect(fakeCloudControllerClient.GetBuildpacksCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetBuildpacksArgsForCall(0)).To(ConsistOf(ccv3.Query{
					Key:    ccv3.OrderBy,
					Values: []string{ccv3.PositionOrder},
				}))
			})
		})
	})

	Describe("CreateBuildpack", func() {
		var (
			buildpack  Buildpack
			warnings   Warnings
			executeErr error
			bp         Buildpack
		)

		JustBeforeEach(func() {
			buildpack, warnings, executeErr = actor.CreateBuildpack(bp)
		})

		When("creating a buildpack fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateBuildpackReturns(
					ccv3.Buildpack{},
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("some-error"))
			})

			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
				Expect(buildpack).To(Equal(Buildpack{}))
			})
		})

		When("creating a buildpack is successful", func() {
			var returnBuildpack Buildpack
			BeforeEach(func() {
				bp = Buildpack{Name: "some-name", Stack: "some-stack"}
				returnBuildpack = Buildpack{GUID: "some-guid", Name: "some-name", Stack: "some-stack"}
				fakeCloudControllerClient.CreateBuildpackReturns(
					ccv3.Buildpack(returnBuildpack),
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					nil)
			})

			It("returns the buildpacks and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
				Expect(buildpack).To(Equal(returnBuildpack))

				Expect(fakeCloudControllerClient.CreateBuildpackCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateBuildpackArgsForCall(0)).To(Equal(ccv3.Buildpack(bp)))
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

})
