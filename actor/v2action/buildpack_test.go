package v2action_test

import (
	"errors"
	"io"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

var _ = Describe("Buildpack", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("CreateBuildpack", func() {
		var (
			buildpack  Buildpack
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			buildpack, warnings, executeErr = actor.CreateBuildpack("some-bp-name", 42, true)
		})

		Context("when creating the buildpack is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateBuildpackReturns(ccv2.Buildpack{GUID: "some-guid"}, ccv2.Warnings{"some-create-warning"}, nil)
			})

			It("returns the buildpack and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.CreateBuildpackCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateBuildpackArgsForCall(0)).To(Equal(ccv2.Buildpack{
					Name:     "some-bp-name",
					Position: 42,
					Enabled:  true,
				}))

				Expect(buildpack).To(Equal(Buildpack{GUID: "some-guid"}))
				Expect(warnings).To(ConsistOf("some-create-warning"))
			})
		})

		Context("when the buildpack already exists with nil stack", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateBuildpackReturns(ccv2.Buildpack{}, ccv2.Warnings{"some-create-warning"}, ccerror.BuildpackAlreadyExistsWithoutStackError{Message: ""})
			})

			It("returns a BuildpackAlreadyExistsWithoutStackError error and all warnings", func() {
				Expect(warnings).To(ConsistOf("some-create-warning"))
				Expect(executeErr).To(MatchError(actionerror.BuildpackAlreadyExistsWithoutStackError("some-bp-name")))
			})
		})

		Context("when the buildpack name is taken", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateBuildpackReturns(ccv2.Buildpack{}, ccv2.Warnings{"some-create-warning"}, ccerror.BuildpackNameTakenError{Message: ""})
			})

			It("returns a BuildpackAlreadyExistsWithoutStackError error and all warnings", func() {
				Expect(warnings).To(ConsistOf("some-create-warning"))
				Expect(executeErr).To(MatchError(actionerror.BuildpackNameTakenError("some-bp-name")))
			})
		})

		Context("when a cc create error occurs", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateBuildpackReturns(ccv2.Buildpack{}, ccv2.Warnings{"some-create-warning"}, errors.New("kaboom"))
			})

			It("returns an error and all warnings", func() {
				Expect(warnings).To(ConsistOf("some-create-warning"))
				Expect(executeErr).To(MatchError("kaboom"))
			})
		})
	})

	Describe("UploadBuildpack", func() {
		var (
			bpFile     io.Reader
			bpFilePath string
			fakePb     *v2actionfakes.FakeSimpleProgressBar

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			bpFile = strings.NewReader("")
		})

		JustBeforeEach(func() {
			fakePb = new(v2actionfakes.FakeSimpleProgressBar)
			fakePb.InitializeReturns(bpFile, 0, nil)
			bpFilePath = "tmp/buildpack.zip"
			warnings, executeErr = actor.UploadBuildpack("some-bp-guid", bpFilePath, fakePb)
		})

		It("tracks the progress of the upload", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakePb.InitializeCallCount()).To(Equal(1))
			Expect(fakePb.InitializeArgsForCall(0)).To(Equal(bpFilePath))
			Expect(fakePb.TerminateCallCount()).To(Equal(1))
		})

		Context("when the upload errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadBuildpackReturns(ccv2.Warnings{"some-upload-warning"}, errors.New("some-upload-error"))
			})

			It("returns warnings and errors", func() {
				Expect(warnings).To(ConsistOf("some-upload-warning"))
				Expect(executeErr).To(MatchError("some-upload-error"))
			})
		})

		Context("when the cc returns an error because the buildpack and stack combo already exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadBuildpackReturns(ccv2.Warnings{"some-upload-warning"}, ccerror.BuildpackAlreadyExistsForStackError{Message: "buildpack stack error"})
			})

			It("returns warnings and a BuildpackAlreadyExistsForStackError", func() {
				Expect(warnings).To(ConsistOf("some-upload-warning"))
				Expect(executeErr).To(MatchError(actionerror.BuildpackAlreadyExistsForStackError{Message: "buildpack stack error"}))
			})
		})

		Context("when the upload is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadBuildpackReturns(ccv2.Warnings{"some-create-warning"}, nil)
			})

			It("uploads the buildpack and returns any warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.UploadBuildpackCallCount()).To(Equal(1))
				guid, path, pbReader, size := fakeCloudControllerClient.UploadBuildpackArgsForCall(0)
				Expect(guid).To(Equal("some-bp-guid"))
				Expect(size).To(Equal(int64(0)))
				Expect(path).To(Equal(bpFilePath))
				Expect(pbReader).To(Equal(bpFile))
				Expect(warnings).To(ConsistOf("some-create-warning"))
			})
		})
	})

	Describe("PrepareBuildpackBits", func() {
		var (
			inPath         string
			outPath        string
			fakeDownloader *v2actionfakes.FakeDownloader

			executeErr error
		)

		BeforeEach(func() {
			fakeDownloader = new(v2actionfakes.FakeDownloader)
		})

		JustBeforeEach(func() {
			outPath, executeErr = actor.PrepareBuildpackBits(inPath, fakeDownloader)
		})

		Context("when the buildpack path is a url", func() {
			BeforeEach(func() {
				inPath = "http://buildpacks.com/a.zip"
				fakeDownloader = new(v2actionfakes.FakeDownloader)
			})

			Context("when downloading the file succeeds", func() {
				BeforeEach(func() {
					fakeDownloader.DownloadReturns("/tmp/a.zip", nil)
				})

				It("downloads the buildpack to a local file", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeDownloader.DownloadCallCount()).To(Equal(1))
					Expect(fakeDownloader.DownloadArgsForCall(0)).To(Equal("http://buildpacks.com/a.zip"))
					Expect(outPath).To(Equal("/tmp/a.zip"))
				})
			})

			Context("when downloading the file fails", func() {
				BeforeEach(func() {
					fakeDownloader.DownloadReturns("", errors.New("bad"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("bad"))
				})
			})
		})

		Context("when the buildpack path points to a file", func() {
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
})
