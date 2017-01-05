package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Space Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		fakeConfig := new(v2actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, nil, fakeConfig)
	})

	Describe("GetOrganizationSpaces", func() {
		Context("when the CC API client does not return any errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns([]ccv2.Space{
					{
						GUID:     "some-space-guid",
						Name:     "some-space",
						AllowSSH: true,
					},
				}, ccv2.Warnings{"get-spaces-warning"}, nil)
			})

			It("returns all warnings and spaces", func() {
				spaces, warnings, err := actor.GetOrganizationSpaces("some-org-guid")

				Expect(spaces).To(ConsistOf(Space{
					GUID:     "some-space-guid",
					Name:     "some-space",
					AllowSSH: true,
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-spaces-warning"))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				expectedQuery := []ccv2.Query{
					{
						Filter:   "organization_guid",
						Operator: ":",
						Value:    "some-org-guid",
					}}
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal(expectedQuery))
			})
		})

		Context("when the CC API client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns([]ccv2.Space{}, ccv2.Warnings{"get-spaces-warning"}, errors.New("cc-get-spaces-error"))
			})

			It("return all warnings and the error", func() {
				spaces, warnings, err := actor.GetOrganizationSpaces("some-org-guid")

				Expect(err).To(MatchError(errors.New("cc-get-spaces-error")))
				Expect(warnings).To(ConsistOf("get-spaces-warning"))
				Expect(spaces).To(Equal([]Space{}))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				expectedQuery := []ccv2.Query{
					{
						Filter:   "organization_guid",
						Operator: ":",
						Value:    "some-org-guid",
					}}
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal(expectedQuery))
			})
		})
	})

	Describe("GetSpaceByName", func() {
		Context("when the CC API client does not return any errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns([]ccv2.Space{
					{
						GUID:     "some-space-guid",
						Name:     "some-space",
						AllowSSH: true,
					},
				}, ccv2.Warnings{"get-spaces-warning"}, nil)
			})

			It("returns all warnings and spaces", func() {
				space, warnings, err := actor.GetSpaceByName("some-org-guid", "some-space")

				Expect(space).To(Equal(Space{
					GUID:     "some-space-guid",
					Name:     "some-space",
					AllowSSH: true,
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-spaces-warning"))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				expectedQuery := []ccv2.Query{
					{
						Filter:   "organization_guid",
						Operator: ":",
						Value:    "some-org-guid",
					},
					{
						Filter:   "name",
						Operator: ":",
						Value:    "some-space",
					},
				}
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(expectedQuery))
			})
		})

		Context("when the CC API client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns([]ccv2.Space{}, ccv2.Warnings{"get-spaces-warning"}, errors.New("cc-get-spaces-error"))
			})

			It("return all warnings and the error", func() {
				space, warnings, err := actor.GetSpaceByName("some-org-guid", "some-space")

				Expect(err).To(MatchError(errors.New("cc-get-spaces-error")))
				Expect(warnings).To(ConsistOf("get-spaces-warning"))
				Expect(space).To(Equal(Space{}))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				expectedQuery := []ccv2.Query{
					{
						Filter:   "organization_guid",
						Operator: ":",
						Value:    "some-org-guid",
					},
					{
						Filter:   "name",
						Operator: ":",
						Value:    "some-space",
					},
				}
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(expectedQuery))
			})
		})

		Context("when the space is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns([]ccv2.Space{}, ccv2.Warnings{"get-spaces-warning"}, nil)
			})

			It("returns all warnings and a SpaceNotFoundError", func() {
				space, warnings, err := actor.GetSpaceByName("some-org-guid", "some-space")

				Expect(err).To(MatchError(SpaceNotFoundError{
					OrgGUID:   "some-org-guid",
					SpaceName: "some-space",
				}))
				Expect(warnings).To(ConsistOf("get-spaces-warning"))
				Expect(space).To(Equal(Space{}))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				expectedQuery := []ccv2.Query{
					{
						Filter:   "organization_guid",
						Operator: ":",
						Value:    "some-org-guid",
					},
					{
						Filter:   "name",
						Operator: ":",
						Value:    "some-space",
					},
				}
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(expectedQuery))
			})
		})

		Context("when multiple spaces are found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns([]ccv2.Space{
					{
						GUID:     "some-space-guid",
						Name:     "some-space",
						AllowSSH: true,
					},
					{
						GUID:     "another-space-guid",
						Name:     "another-space",
						AllowSSH: true,
					},
				}, ccv2.Warnings{"get-spaces-warning"}, nil)
			})

			It("returns all warnings and a MultipleSpacesFoundError", func() {
				space, warnings, err := actor.GetSpaceByName("some-org-guid", "some-space")

				Expect(err).To(MatchError(MultipleSpacesFoundError{
					OrgGUID:   "some-org-guid",
					SpaceName: "some-space",
				}))
				Expect(warnings).To(ConsistOf("get-spaces-warning"))
				Expect(space).To(Equal(Space{}))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				expectedQuery := []ccv2.Query{
					{
						Filter:   "organization_guid",
						Operator: ":",
						Value:    "some-org-guid",
					},
					{
						Filter:   "name",
						Operator: ":",
						Value:    "some-space",
					},
				}
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(expectedQuery))
			})
		})
	})
})
