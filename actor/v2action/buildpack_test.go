package v2action_test

import (
	"errors"
	"io"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
				Expect(fakeCloudControllerClient.CreateBuildpackCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CreateBuildpackArgsForCall(0)).To(Equal(ccv2.Buildpack{
					Name:     "some-bp-name",
					Position: 42,
					Enabled:  true,
				}))

				Expect(buildpack).To(Equal(Buildpack{GUID: "some-guid"}))
				Expect(warnings).To(ConsistOf("some-create-warning"))
				Expect(executeErr).ToNot(HaveOccurred())
			})
		})

		Context("when the buildpack already exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateBuildpackReturns(ccv2.Buildpack{}, ccv2.Warnings{"some-create-warning"}, ccerror.BuildpackAlreadyExistsError{Message: ""})
			})

			It("returns a BuildpackAlreadyExistsError error and all warnings", func() {
				Expect(warnings).To(ConsistOf("some-create-warning"))
				Expect(executeErr).To(MatchError("A buildpack with the name some-bp-name already exists"))
			})
		})

		Context("when a cc create error occurs", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateBuildpackReturns(ccv2.Buildpack{}, ccv2.Warnings{"some-create-warning"}, errors.New("some error oh no"))
			})

			It("returns a BuildpackAlreadyExistsError error and all warnings", func() {
				Expect(warnings).To(ConsistOf("some-create-warning"))
				Expect(executeErr).To(MatchError("some error oh no"))
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
			bpFilePath = "some/fake-buildpack.zip"
		})

		JustBeforeEach(func() {
			fakePb = new(v2actionfakes.FakeSimpleProgressBar)
			fakePb.InitializeReturns(bpFile, 0, nil)
			warnings, executeErr = actor.UploadBuildpack("some-bp-guid", bpFilePath, fakePb)
		})

		It("tracks the progress of the upload", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakePb.InitializeCallCount()).To(Equal(1))
			Expect(fakePb.InitializeArgsForCall(0)).To(Equal(bpFilePath))
			Expect(fakePb.TerminateCallCount()).To(Equal(1))
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

		Context("when a cc upload error occurs", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadBuildpackReturns(ccv2.Warnings{"some-upload-warning"}, errors.New("some-upload-error"))
			})

			It("returns warnings and errors", func() {
				Expect(warnings).To(ConsistOf("some-upload-warning"))
				Expect(executeErr).To(MatchError("some-upload-error"))
			})
		})
	})
})
