package v2action_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

var _ = Describe("Service Access", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		warnings                  ccv2.Warnings
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("EnablePlanForAllOrgs", func() {
		var (
			enablePlanErr      error
			enablePlanWarnings Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServicesReturns(
				[]ccv2.Service{
					{Label: "service-1", GUID: "service-guid-1"},
				},
				warnings,
				nil)

			fakeCloudControllerClient.GetServicePlansReturns(
				[]ccv2.ServicePlan{
					{Name: "plan-1", GUID: "service-plan-guid-1"},
					{Name: "plan-2", GUID: "service-plan-guid-2"},
				},
				warnings,
				nil)
		})

		JustBeforeEach(func() {
			enablePlanWarnings, enablePlanErr = actor.EnablePlanForAllOrgs("service-1", "plan-2")
		})

		It("updates the service plan visibility", func() {
			Expect(enablePlanErr).NotTo(HaveOccurred())
			Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.UpdateServicePlanCallCount()).To(Equal(1))

			planGuid, visible := fakeCloudControllerClient.UpdateServicePlanArgsForCall(0)
			Expect(planGuid).To(Equal("service-plan-guid-2"))
			Expect(visible).To(BeTrue())
		})

		When("the plan is already visible in some orgs", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlanVisibilitiesReturns(
					[]ccv2.ServicePlanVisibility{
						{GUID: "service-visibility-guid-1"},
						{GUID: "service-visibility-guid-2"},
					},
					warnings,
					nil)
			})

			It("first disables the plan in both orgs", func() {
				Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesCallCount()).To(Equal(1))

				filters := fakeCloudControllerClient.GetServicePlanVisibilitiesArgsForCall(0)
				Expect(filters[0].Type).To(Equal(constant.ServicePlanGUIDFilter))
				Expect(filters[0].Operator).To(Equal(constant.EqualOperator))
				Expect(filters[0].Values).To(Equal([]string{"service-plan-guid-2"}))

				Expect(fakeCloudControllerClient.DeleteServicePlanVisibilityCallCount()).To(Equal(2))
			})

			When("getting service plan visibilities fails", func() {
				var expectedError error

				BeforeEach(func() {
					expectedError = fmt.Errorf("oh no")

					fakeCloudControllerClient.GetServicePlanVisibilitiesReturns(
						[]ccv2.ServicePlanVisibility{},
						warnings,
						expectedError)
				})

				It("returns the error", func() {
					Expect(enablePlanErr).To(MatchError(expectedError))
				})
			})
		})

		When("getting service plans fails", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = fmt.Errorf("it didn't work!")
				fakeCloudControllerClient.GetServicePlansReturns(nil, warnings, expectedError)
			})

			It("returns the error", func() {
				Expect(enablePlanErr).To(MatchError(expectedError))
			})
		})

		When("getting services fails", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = fmt.Errorf("very bad")
				fakeCloudControllerClient.GetServicesReturns(nil, warnings, expectedError)
			})

			It("returns the error", func() {
				Expect(enablePlanErr).To(MatchError(expectedError))
			})
		})

		When("there are no matching services", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{}, []string{"warning-1"}, nil)
			})

			It("returns not found error", func() {
				Expect(enablePlanErr).To(MatchError(actionerror.ServiceNotFoundError{Name: "service-1"}))
			})

			It("returns all warnings", func() {
				Expect(enablePlanWarnings).To(ConsistOf([]string{"warning-1"}))
			})
		})

		When("there are no matching plans", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns([]ccv2.ServicePlan{}, []string{"warning-1"}, nil)
			})

			It("returns not found error", func() {
				Expect(enablePlanErr).To(MatchError(actionerror.ServicePlanNotFoundError{Name: "plan-2", ServiceName: "service-1"}))
			})

			It("returns all warnings", func() {
				Expect(enablePlanWarnings).To(ConsistOf([]string{"warning-1"}))
			})
		})

		When("updating service plan fails", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = fmt.Errorf("the worst")
				fakeCloudControllerClient.UpdateServicePlanReturns(warnings, expectedError)
			})

			It("returns the error", func() {
				Expect(enablePlanErr).To(MatchError(expectedError))
			})
		})

		When("there are warnings", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{
						{Label: "service-1", GUID: "service-guid-1"},
					},
					[]string{"foo"},
					nil)

				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv2.ServicePlan{
						{Name: "plan-1", GUID: "service-plan-guid-1"},
						{Name: "plan-2", GUID: "service-plan-guid-2"},
					},
					[]string{"bar"},
					nil)

				fakeCloudControllerClient.GetServicePlanVisibilitiesReturns(
					[]ccv2.ServicePlanVisibility{
						{GUID: "service-visibility-guid-1"},
						{GUID: "service-visibility-guid-2"},
					},
					[]string{"baz"},
					nil)

				fakeCloudControllerClient.UpdateServicePlanReturns(
					[]string{"qux"},
					nil)
			})

			It("returns them", func() {
				Expect(enablePlanWarnings).To(ConsistOf([]string{"foo", "bar", "baz", "qux"}))
			})

			When("getting services fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicesReturns(nil, []string{"foo"}, fmt.Errorf("err"))
				})

				It("still returns warnings", func() {
					Expect(enablePlanWarnings).To(ConsistOf([]string{"foo"}))
				})
			})

			When("getting service plans fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns(nil, []string{"bar"}, fmt.Errorf("err"))
				})

				It("still returns warnings", func() {
					Expect(enablePlanWarnings).To(ConsistOf([]string{"foo", "bar"}))
				})
			})

			When("getting service plan visibilities fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlanVisibilitiesReturns(nil, []string{"baz"}, fmt.Errorf("err"))
				})

				It("still returns warnings", func() {
					Expect(enablePlanWarnings).To(ConsistOf([]string{"foo", "bar", "baz"}))
				})
			})

			When("update service plan fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateServicePlanReturns([]string{"qux"}, fmt.Errorf("err"))
				})

				It("still returns warnings", func() {
					Expect(enablePlanWarnings).To(ConsistOf([]string{"foo", "bar", "baz", "qux"}))
				})
			})
		})
	})

	Describe("EnablePlanForOrg", func() {
		var (
			enablePlanWarnings Warnings
			enablePlanErr      error
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServicesReturns(
				[]ccv2.Service{
					{Label: "service-1", GUID: "service-guid-1"},
				},
				nil, nil)

			fakeCloudControllerClient.GetServicePlansReturns(
				[]ccv2.ServicePlan{
					{Name: "plan-1", GUID: "service-plan-guid-1"},
					{Name: "plan-2", GUID: "service-plan-guid-2"},
				},
				nil, nil)
		})

		JustBeforeEach(func() {
			enablePlanWarnings, enablePlanErr = actor.EnablePlanForOrg("service-1", "plan-2", "my-org")
		})

		When("the specified service does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{},
					nil, nil)
			})

			It("filters by service name and returns service not found error", func() {
				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(enablePlanErr).To(MatchError(actionerror.ServiceNotFoundError{Name: "service-1"}))
				filters := fakeCloudControllerClient.GetServicesArgsForCall(0)
				Expect(len(filters)).To(Equal(1))
				Expect(filters[0]).To(Equal(ccv2.Filter{
					Type:     constant.LabelFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"service-1"},
				}))
			})
		})

		When("the specified org exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{
						{Name: "my-org", GUID: "org-guid-1"},
					}, nil, nil)
			})

			It("enables the plan", func() {
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))

				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				serviceFilter := fakeCloudControllerClient.GetServicePlansArgsForCall(0)
				Expect(serviceFilter[0].Values).To(ContainElement("service-guid-1"))

				Expect(fakeCloudControllerClient.CreateServicePlanVisibilityCallCount()).To(Equal(1))
				planGUID, orgGUID := fakeCloudControllerClient.CreateServicePlanVisibilityArgsForCall(0)
				Expect(planGUID).To(Equal("service-plan-guid-2"))
				Expect(orgGUID).To(Equal("org-guid-1"))

				Expect(enablePlanErr).NotTo(HaveOccurred())
			})

			When("warnings are raised", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateServicePlanVisibilityReturns(ccv2.ServicePlanVisibility{}, []string{"foo"}, nil)
				})

				It("still returns warnings", func() {
					Expect(enablePlanWarnings).To(ConsistOf([]string{"foo"}))
				})
			})

			When("the specified plan does exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns(
						[]ccv2.ServicePlan{
							{Name: "plan-4", GUID: "service-plan-guid-4"},
							{Name: "plan-5", GUID: "service-plan-guid-5"},
						},
						nil, nil)
				})

				It("returns an error", func() {
					Expect(enablePlanErr.Error()).To(Equal("Service plan 'plan-2' not found"))
				})
			})

			When("getting services fails", func() {
				var expectedError error

				BeforeEach(func() {
					expectedError = fmt.Errorf("it broke")
					fakeCloudControllerClient.GetServicesReturns(nil, []string{"foo"}, expectedError)
				})

				It("returns the error", func() {
					Expect(enablePlanErr).To(MatchError(expectedError))
				})

				It("still returns warnings", func() {
					Expect(enablePlanWarnings).To(ConsistOf([]string{"foo"}))
				})
			})

			When("getting service plans fails", func() {
				var expectedError error

				BeforeEach(func() {
					fakeCloudControllerClient.GetServicesReturns(
						[]ccv2.Service{
							{Label: "service-1", GUID: "service-guid-1"},
						},
						[]string{"foo"},
						nil)

					expectedError = fmt.Errorf("it broke")
					fakeCloudControllerClient.GetServicePlansReturns(nil, []string{"bar"}, expectedError)
				})

				It("still returns warnings", func() {
					Expect(enablePlanWarnings).To(ConsistOf([]string{"foo", "bar"}))
				})

				It("returns the error", func() {
					Expect(enablePlanErr).To(MatchError(expectedError))
				})
			})

			When("create service plan visibility fails", func() {
				var expectedError error

				BeforeEach(func() {
					fakeCloudControllerClient.GetServicesReturns(
						[]ccv2.Service{
							{Label: "service-1", GUID: "service-guid-1"},
						},
						[]string{"foo"},
						nil)

					fakeCloudControllerClient.GetServicePlansReturns(
						[]ccv2.ServicePlan{
							{Name: "plan-1", GUID: "service-plan-guid-1"},
							{Name: "plan-2", GUID: "service-plan-guid-2"},
						},
						[]string{"bar"},
						nil)

					expectedError = fmt.Errorf("some error")
					fakeCloudControllerClient.CreateServicePlanVisibilityReturns(ccv2.ServicePlanVisibility{}, []string{"baz"}, expectedError)
				})

				It("still returns warnings", func() {
					Expect(enablePlanWarnings).To(ConsistOf([]string{"foo", "bar", "baz"}))
				})

				It("returns the error", func() {
					Expect(enablePlanErr).To(MatchError(expectedError))
				})
			})
		})

		When("the specified org does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv2.Organization{}, nil, nil)
			})

			It("filters by the organization name and returns an organization not found error", func() {
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				Expect(enablePlanErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "my-org"}))
				filters := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(len(filters)).To(Equal(1))
				Expect(filters[0]).To(Equal(ccv2.Filter{
					Type:     constant.NameFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"my-org"},
				}))
			})
		})
	})

	Describe("EnableServiceForOrg", func() {
		var enableServiceForOrgErr error
		var enableServiceForOrgWarnings Warnings

		JustBeforeEach(func() {
			enableServiceForOrgWarnings, enableServiceForOrgErr = actor.EnableServiceForOrg("service-1", "my-org")
		})

		When("the service does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{},
					nil, nil)
			})

			It("filters by service name and returns service not found error", func() {
				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(enableServiceForOrgErr).To(MatchError(actionerror.ServiceNotFoundError{Name: "service-1"}))

				filters := fakeCloudControllerClient.GetServicesArgsForCall(0)
				Expect(len(filters)).To(Equal(1))
				Expect(filters[0]).To(Equal(ccv2.Filter{
					Type:     constant.LabelFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"service-1"},
				}))
			})
		})

		When("the specified org does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{{Label: "service-1", GUID: "service-guid-1"}},
					nil, nil)
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv2.Organization{}, nil, nil)
			})

			It("filters by the organization name and returns an organization not found error", func() {
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				Expect(enableServiceForOrgErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "my-org"}))
				filters := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(len(filters)).To(Equal(1))
				Expect(filters[0]).To(Equal(ccv2.Filter{
					Type:     constant.NameFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"my-org"},
				}))
			})
		})

		When("the service and org exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{{Label: "service-1", GUID: "service-guid-1"}},
					nil, nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{Name: "my-org", GUID: "org-guid-1"}},
					nil, nil)
				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv2.ServicePlan{
						{Name: "plan-1", GUID: "service-plan-guid-1"},
						{Name: "plan-2", GUID: "service-plan-guid-2"},
					},
					warnings,
					nil)
			})

			It("enables all the plans for that org", func() {
				Expect(enableServiceForOrgErr).NotTo(HaveOccurred())
				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))

				Expect(fakeCloudControllerClient.CreateServicePlanVisibilityCallCount()).To(Equal(2))
				planGUID1, orgGUID1 := fakeCloudControllerClient.CreateServicePlanVisibilityArgsForCall(0)
				planGUID2, orgGUID2 := fakeCloudControllerClient.CreateServicePlanVisibilityArgsForCall(1)

				Expect(orgGUID1).To(Equal("org-guid-1"))
				Expect(orgGUID2).To(Equal("org-guid-1"))
				Expect([]string{planGUID1, planGUID2}).To(ConsistOf([]string{"service-plan-guid-1", "service-plan-guid-2"}))
			})
		})

		When("getting services fails", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = fmt.Errorf("this is very bad")
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{},
					ccv2.Warnings{"foo"},
					expectedError)
			})

			It("returns the error", func() {
				Expect(enableServiceForOrgErr).To(MatchError(expectedError))
			})

			It("still returns warnings", func() {
				Expect(enableServiceForOrgWarnings).To(ConsistOf("foo"))
			})
		})

		When("getting organizations fails", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = fmt.Errorf("get-orgs-error")
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{{Label: "service-1", GUID: "service-guid-1"}},
					ccv2.Warnings{"foo"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{Name: "my-org", GUID: "org-guid-1"}},
					ccv2.Warnings{"bar"},
					expectedError)
			})

			It("returns the error", func() {
				Expect(enableServiceForOrgErr).To(MatchError(expectedError))
			})

			It("still returns warnings", func() {
				Expect(enableServiceForOrgWarnings).To(ConsistOf("foo", "bar"))
			})
		})

		When("getting service plans fails", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = fmt.Errorf("get-service-plans-error")
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{{Label: "service-1", GUID: "service-guid-1"}},
					ccv2.Warnings{"foo"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{Name: "my-org", GUID: "org-guid-1"}},
					ccv2.Warnings{"bar"},
					nil)
				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv2.ServicePlan{},
					ccv2.Warnings{"baz"},
					expectedError)
			})

			It("returns the error", func() {
				Expect(enableServiceForOrgErr).To(MatchError(expectedError))
			})

			It("still returns warnings", func() {
				Expect(enableServiceForOrgWarnings).To(ConsistOf("foo", "bar", "baz"))
			})
		})

		When("creating service plan visibility fails", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = fmt.Errorf("create-service-plan-vis-error")
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{{Label: "service-1", GUID: "service-guid-1"}},
					ccv2.Warnings{"foo"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{Name: "my-org", GUID: "org-guid-1"}},
					ccv2.Warnings{"bar"},
					nil)
				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv2.ServicePlan{
						{Name: "service-plan-1", GUID: "service-plan-guid-1"}},
					ccv2.Warnings{"baz"},
					nil)
				fakeCloudControllerClient.CreateServicePlanVisibilityReturns(
					ccv2.ServicePlanVisibility{},
					ccv2.Warnings{"qux"},
					expectedError)
			})

			It("returns the error", func() {
				Expect(enableServiceForOrgErr).To(MatchError(expectedError))
			})

			It("still returns warnings", func() {
				Expect(enableServiceForOrgWarnings).To(ConsistOf("foo", "bar", "baz", "qux"))
			})
		})
	})

	Describe("EnableServiceForAllOrgs", func() {
		var (
			enableServiceErr      error
			enableServiceWarnings Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServicesReturns(
				[]ccv2.Service{
					{Label: "service-1", GUID: "service-guid-1"},
				},
				warnings,
				nil)

			fakeCloudControllerClient.GetServicePlansReturns(
				[]ccv2.ServicePlan{
					{Name: "plan-1", GUID: "service-plan-guid-1"},
					{Name: "plan-2", GUID: "service-plan-guid-2"},
				},
				warnings,
				nil)
		})

		JustBeforeEach(func() {
			enableServiceWarnings, enableServiceErr = actor.EnableServiceForAllOrgs("service-1")
		})

		It("should update all plans to public", func() {
			Expect(enableServiceErr).NotTo(HaveOccurred())
			Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.UpdateServicePlanCallCount()).To(Equal(2))
			planGuid, public := fakeCloudControllerClient.UpdateServicePlanArgsForCall(0)
			Expect(planGuid).To(Equal("service-plan-guid-1"))
			Expect(public).To(BeTrue())
			planGuid, public = fakeCloudControllerClient.UpdateServicePlanArgsForCall(1)
			Expect(planGuid).To(Equal("service-plan-guid-2"))
			Expect(public).To(BeTrue())
		})

		When("the service does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{},
					nil, nil)
			})

			It("filters by service name and returns service not found error", func() {
				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(enableServiceErr).To(MatchError(actionerror.ServiceNotFoundError{Name: "service-1"}))

				filters := fakeCloudControllerClient.GetServicesArgsForCall(0)
				Expect(len(filters)).To(Equal(1))
				Expect(filters[0]).To(Equal(ccv2.Filter{
					Type:     constant.LabelFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"service-1"},
				}))
			})
		})

		When("getting services fails", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = fmt.Errorf("it broke")
				fakeCloudControllerClient.GetServicesReturns(nil, []string{"foo"}, expectedError)
			})

			It("returns the error", func() {
				Expect(enableServiceErr).To(MatchError(expectedError))
			})

			It("still returns warnings", func() {
				Expect(enableServiceWarnings).To(ConsistOf([]string{"foo"}))
			})
		})

		When("getting service plans fails", func() {
			var expectedError error

			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{
						{Label: "service-1", GUID: "service-guid-1"},
					},
					[]string{"foo"},
					nil)

				expectedError = fmt.Errorf("it broke")
				fakeCloudControllerClient.GetServicePlansReturns(nil, []string{"bar"}, expectedError)
			})

			It("still returns warnings", func() {
				Expect(enableServiceWarnings).To(ConsistOf([]string{"foo", "bar"}))
			})

			It("returns the error", func() {
				Expect(enableServiceErr).To(MatchError(expectedError))
			})
		})

		When("update service plan fails", func() {
			var expectedError error

			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{
						{Label: "service-1", GUID: "service-guid-1"},
					},
					[]string{"foo"},
					nil)

				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv2.ServicePlan{
						{Name: "plan-1", GUID: "service-plan-guid-1"},
						{Name: "plan-2", GUID: "service-plan-guid-2"},
					},
					[]string{"bar"},
					nil)

				expectedError = fmt.Errorf("some error")
				fakeCloudControllerClient.UpdateServicePlanReturns([]string{"baz"}, expectedError)
			})

			It("still returns warnings", func() {
				Expect(enableServiceWarnings).To(ConsistOf([]string{"foo", "bar", "baz"}))
			})

			It("returns the error", func() {
				Expect(enableServiceErr).To(MatchError(expectedError))
			})
		})
	})
})
