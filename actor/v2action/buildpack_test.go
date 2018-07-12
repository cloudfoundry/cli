package v2action_test

import (
	"errors"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

var _ = FDescribe("Buildpack", func() {
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

		Context("when the cloud controller client is successful", func() {
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

		Context("when a cc error occurs", func() {
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
			fakePb *v2actionfakes.FakeProgressBar

			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			fakePb = new(v2actionfakes.FakeProgressBar)
			warnings, executeErr = actor.UploadBuildpack("some-bp-guid", "some-path", fakePb)
		})

		It("tracks the progress of the upload", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakePb.InitializeCallCount()).To(Equal(1))
			Expect(fakePb.InitializeArgsForCall(0)).To(Equal("some-path"))
			Expect(fakePb.TerminateCallCount()).To(Equal(1))
		})

		Context("when the upload is successful", func() {
			var buildpackFile *os.File

			BeforeEach(func() {
				var err error
				buildpackFile, err = ioutil.TempFile("", "test-buildpack")
				Expect(err).ToNot(HaveOccurred())
				buildpackFile.Close()
				err = ioutil.WriteFile(buildpackFile.Name(), []byte("123456"), 0655)
				Expect(err).ToNot(HaveOccurred())
				fakeCloudControllerClient.UploadBuildpackReturns(ccv2.Warnings{"some-create-warning"}, nil)
			})

			AfterEach(func() {
				Expect(os.RemoveAll(buildpackFile.Name()))
			})

			It("uploads the buildpack and returns any warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.UploadBuildpackCallCount()).To(Equal(1))
				guid, _, size := fakeCloudControllerClient.UploadBuildpackArgsForCall(0)
				Expect(guid).To(Equal("some-bp-guid"))
				Expect(size).To(Equal(6))
				Expect(warnings).To(ConsistOf("some-create-warning"))
			})
		})
	})
})
