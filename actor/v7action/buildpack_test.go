package v7action_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"os"
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
					{Name: "buildpack-1", Position: 1},
					{Name: "buildpack-2", Position: 2},
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
					{Name: "buildpack-1", Position: 1},
					{Name: "buildpack-2", Position: 2},
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
			warnings, executeErr = actor.UploadBuildpack("some-bp-guid", bpFilePath, fakePb)
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
				fakeCloudControllerClient.UploadBuildpackReturns(ccv3.Warnings{"some-upload-warning"}, errors.New("some-upload-error"))
			})

			It("returns warnings and errors", func() {
				Expect(warnings).To(ConsistOf("some-upload-warning"))
				Expect(executeErr).To(MatchError("some-upload-error"))
			})
		})

		When("the cc returns an error because the buildpack and stack combo already exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadBuildpackReturns(ccv3.Warnings{"some-upload-warning"}, ccerror.BuildpackAlreadyExistsForStackError{Message: "ya blew it"})
			})

			It("returns warnings and a BuildpackAlreadyExistsForStackError", func() {
				Expect(warnings).To(ConsistOf("some-upload-warning"))
				Expect(executeErr).To(MatchError(actionerror.BuildpackAlreadyExistsForStackError{Message: "ya blew it"}))
			})
		})

		When("the upload is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UploadBuildpackReturns(
					ccv3.Warnings{"some-create-warning"},
					nil,
				)
			})

			It("uploads the buildpack and returns any warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.UploadBuildpackCallCount()).To(Equal(1))
				guid, path, pbReader, size := fakeCloudControllerClient.UploadBuildpackArgsForCall(0)
				Expect(guid).To(Equal("some-bp-guid"))
				Expect(size).To(Equal(int64(66)))
				Expect(path).To(Equal(bpFilePath))
				Expect(pbReader).To(Equal(bpFile))
				Expect(warnings).To(ConsistOf("some-create-warning"))
			})
		})
	})

})
