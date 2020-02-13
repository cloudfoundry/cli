package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("space features", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeConfig := new(v7actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil, nil)
	})

	Describe("GetSpaceFeature", func() {
		var (
			spaceName   string
			orgGUID     string
			featureName string
			enabled     bool
			warnings    Warnings
			executeErr  error
		)

		BeforeEach(func() {
			spaceName = "some-space-name"
			orgGUID = "some-org-guid"
			featureName = "ssh"

			fakeCloudControllerClient.GetSpacesReturns(
				[]ccv3.Space{
					ccv3.Space{
						Name: spaceName,
						GUID: "some-space-guid",
					},
				},
				ccv3.Warnings{"get-space-warning"},
				nil,
			)

			fakeCloudControllerClient.GetSpaceFeatureReturns(
				true,
				ccv3.Warnings{"get-space-feature-warning"},
				nil,
			)
		})

		JustBeforeEach(func() {
			enabled, warnings, executeErr = actor.GetSpaceFeature(spaceName, orgGUID, featureName)
		})

		It("finds the space and retrieves the feature value", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
			query := fakeCloudControllerClient.GetSpacesArgsForCall(0)
			Expect(query).To(ConsistOf(
				ccv3.Query{Key: ccv3.NameFilter, Values: []string{spaceName}},
				ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
			))

			Expect(fakeCloudControllerClient.GetSpaceFeatureCallCount()).To(Equal(1))
			inputSpaceGUID, inputFeature := fakeCloudControllerClient.GetSpaceFeatureArgsForCall(0)
			Expect(inputSpaceGUID).To(Equal("some-space-guid"))
			Expect(inputFeature).To(Equal(featureName))

			Expect(enabled).To(BeTrue())
			Expect(warnings).To(ConsistOf("get-space-warning", "get-space-feature-warning"))
		})

		Context("when the space does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv3.Space{},
					ccv3.Warnings{"get-space-warning"},
					nil,
				)
			})

			It("returns a SpaceNotFoundError", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: spaceName}))
				Expect(warnings).To(ConsistOf("get-space-warning"))
			})
		})

		Context("when an API error occurs", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceFeatureReturns(
					true,
					ccv3.Warnings{"get-space-feature-warning"},
					errors.New("space-feature-error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError("space-feature-error"))
				Expect(warnings).To(ConsistOf("get-space-warning", "get-space-feature-warning"))
			})
		})
	})

	Describe("AllowSpaceSSH", func() {

		var (
			spaceName  string
			spaceGUID  string
			orgGUID    string
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			spaceName = "some-space-name"
			spaceGUID = "some-space-guid"
			orgGUID = "some-org-guid"
			fakeCloudControllerClient.GetSpacesReturns(
				[]ccv3.Space{
					ccv3.Space{
						Name: spaceName,
						GUID: spaceGUID,
					},
				},
				ccv3.Warnings{"get-space-warning"},
				nil,
			)

			fakeCloudControllerClient.UpdateSpaceFeatureReturns(
				ccv3.Warnings{"update-space-feature-warning"},
				nil,
			)
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.AllowSpaceSSH(spaceName, orgGUID)
		})

		When("the GetSpaceFeature returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceFeatureReturns(
					false,
					ccv3.Warnings{"get-space-feature-warning"},
					errors.New("get-ssh-error"),
				)
			})

			It("returns a the error", func() {
				Expect(executeErr).To(MatchError("get-ssh-error"))
			})

		})

		When("space ssh is already enabled", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceFeatureReturns(
					true,
					ccv3.Warnings{"get-space-feature-warning"},
					nil,
				)
			})

			It("returns a descriptive error", func() {
				Expect(executeErr).To(MatchError(actionerror.SpaceSSHAlreadyEnabledError{
					Space: spaceName}))
			})

		})

		When("the update space feature returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateSpaceFeatureReturns(
					ccv3.Warnings{"get-space-feature-warning"},
					errors.New("update-ssh-error"),
				)
			})

			It("returns a the error", func() {
				Expect(executeErr).To(MatchError("update-ssh-error"))
			})

		})

		It("enables the ssh feature for the space", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
			query := fakeCloudControllerClient.GetSpacesArgsForCall(0)
			Expect(query).To(ConsistOf(
				ccv3.Query{Key: ccv3.NameFilter, Values: []string{spaceName}},
				ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
			))

			Expect(fakeCloudControllerClient.UpdateSpaceFeatureCallCount()).To(Equal(1))

			spaceGUID, enable, feature := fakeCloudControllerClient.UpdateSpaceFeatureArgsForCall(0)

			Expect(spaceGUID).To(Equal(spaceGUID))
			Expect(enable).To(Equal(true))
			Expect(feature).To(Equal("ssh"))
			Expect(warnings).To(ConsistOf("get-space-warning", "update-space-feature-warning"))

		})
	})
})
