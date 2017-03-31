package v3action_test

import (
	"errors"
	"net/url"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Organization Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil)
	})

	Describe("GetOrganizationByName", func() {
		Context("when the org exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv3.Organization{
						{
							Name: "some-org-name",
							GUID: "some-org-guid",
						},
					},
					ccv3.Warnings{"some-warning"},
					nil,
				)
			})

			It("returns the organization and warnings", func() {
				org, warnings, err := actor.GetOrganizationByName("some-org-name")
				Expect(err).ToNot(HaveOccurred())
				Expect(org).To(Equal(Organization{
					Name: "some-org-name",
					GUID: "some-org-guid",
				}))
				Expect(warnings).To(Equal(Warnings{"some-warning"}))

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				expectedQuery := url.Values{
					ccv3.NameFilter: []string{"some-org-name"},
				}
				query := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(query).To(Equal(expectedQuery))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv3.Organization{},
					ccv3.Warnings{"some-warning"},
					expectedError)
			})

			It("returns the warnings and the error", func() {
				_, warnings, err := actor.GetOrganizationByName("some-org-name")
				Expect(warnings).To(ConsistOf("some-warning"))
				Expect(err).To(MatchError(expectedError))
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				expectedQuery := url.Values{
					ccv3.NameFilter: []string{"some-org-name"},
				}
				query := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(query).To(Equal(expectedQuery))
			})
		})
	})

	Context("when the org does not exist", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.GetOrganizationsReturns(
				[]ccv3.Organization{},
				ccv3.Warnings{"some-warning"},
				nil,
			)
		})

		It("returns an OrganizationNotFoundError and the warnings", func() {
			_, warnings, err := actor.GetOrganizationByName("some-org-name")
			Expect(warnings).To(ConsistOf("some-warning"))
			Expect(err).To(MatchError(OrganizationNotFoundError{Name: "some-org-name"}))
			Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
			expectedQuery := url.Values{
				ccv3.NameFilter: []string{"some-org-name"},
			}
			query := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
			Expect(query).To(Equal(expectedQuery))
		})
	})
})
