package v2action_test

import (
	"errors"

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
				nil, nil)

			fakeCloudControllerClient.GetServicePlansReturns(
				[]ccv2.ServicePlan{
					{Name: "plan-1", GUID: "service-plan-guid-1"},
					{Name: "plan-2", GUID: "service-plan-guid-2"},
				},
				nil, nil)
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
					nil, nil)
			})

			It("disables the plan in both orgs", func() {
				Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesCallCount()).To(Equal(1))

				filters := fakeCloudControllerClient.GetServicePlanVisibilitiesArgsForCall(0)
				Expect(filters[0].Type).To(Equal(constant.ServicePlanGUIDFilter))
				Expect(filters[0].Operator).To(Equal(constant.EqualOperator))
				Expect(filters[0].Values).To(Equal([]string{"service-plan-guid-2"}))

				Expect(fakeCloudControllerClient.DeleteServicePlanVisibilityCallCount()).To(Equal(2))
			})

			It("and updates the service plan visibility for all orgs", func() {
				Expect(enablePlanErr).NotTo(HaveOccurred())
				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateServicePlanCallCount()).To(Equal(1))

				planGuid, visible := fakeCloudControllerClient.UpdateServicePlanArgsForCall(0)
				Expect(planGuid).To(Equal("service-plan-guid-2"))
				Expect(visible).To(BeTrue())
			})

			When("deleting service plan visibilities fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteServicePlanVisibilityReturns(
						[]string{"visibility-warning"}, errors.New("delete-visibility-error"))
				})

				It("propagates the error", func() {
					Expect(enablePlanErr).To(MatchError(errors.New("delete-visibility-error")))
				})

				It("returns the warnings", func() {
					Expect(enablePlanWarnings).To(Equal(Warnings{"visibility-warning"}))
				})
			})
		})

		When("getting service plan visibilities fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlanVisibilitiesReturns(
					[]ccv2.ServicePlanVisibility{},
					[]string{"a-warning", "another-warning"},
					errors.New("oh no"))
			})

			It("returns the error", func() {
				Expect(enablePlanErr).To(MatchError(errors.New("oh no")))
			})

			It("returns the warnings", func() {
				Expect(enablePlanWarnings).To(Equal(Warnings{"a-warning", "another-warning"}))
			})
		})

		When("getting service plans fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns(nil,
					[]string{"a-warning", "another-warning"},
					errors.New("it didn't work!"))
			})

			It("returns the error", func() {
				Expect(enablePlanErr).To(MatchError(errors.New("it didn't work!")))
			})

			It("returns the warnings", func() {
				Expect(enablePlanWarnings).To(Equal(Warnings{"a-warning", "another-warning"}))
			})
		})

		When("getting services fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(nil,
					[]string{"a-warning", "another-warning"},
					errors.New("it didn't work!"))
			})

			It("returns the error", func() {
				Expect(enablePlanErr).To(MatchError(errors.New("it didn't work!")))
			})

			It("returns the warnings", func() {
				Expect(enablePlanWarnings).To(Equal(Warnings{"a-warning", "another-warning"}))
			})
		})

		When("there are no matching services", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{}, []string{"warning-1", "warning-2"}, nil)
			})

			It("returns not found error", func() {
				Expect(enablePlanErr).To(MatchError(actionerror.ServiceNotFoundError{Name: "service-1"}))
			})

			It("returns all warnings", func() {
				Expect(enablePlanWarnings).To(ConsistOf([]string{"warning-1", "warning-2"}))
			})
		})

		When("there are no matching plans", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns([]ccv2.ServicePlan{}, []string{"warning-1", "warning-2"}, nil)
			})

			It("returns not found error", func() {
				Expect(enablePlanErr).To(MatchError(actionerror.ServicePlanNotFoundError{PlanName: "plan-2", ServiceName: "service-1"}))
			})

			It("returns all warnings", func() {
				Expect(enablePlanWarnings).To(ConsistOf([]string{"warning-1", "warning-2"}))
			})
		})

		When("updating service plan fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateServicePlanReturns(
					[]string{"a-warning", "another-warning"},
					errors.New("it didn't work!"))
			})

			It("returns the error", func() {
				Expect(enablePlanErr).To(MatchError(errors.New("it didn't work!")))
			})

			It("returns the warnings", func() {
				Expect(enablePlanWarnings).To(Equal(Warnings{"a-warning", "another-warning"}))
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

			It("returns the warnings", func() {
				Expect(enablePlanWarnings).To(ConsistOf([]string{"foo", "bar", "baz", "qux"}))
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

			It("returns service not found error", func() {
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
					fakeCloudControllerClient.CreateServicePlanVisibilityReturns(ccv2.ServicePlanVisibility{}, []string{"foo", "bar"}, nil)
				})

				It("returns all warnings", func() {
					Expect(enablePlanWarnings).To(ConsistOf([]string{"foo", "bar"}))
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
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicesReturns(nil, []string{"foo", "bar"}, errors.New("it broke"))
				})

				It("returns the error", func() {
					Expect(enablePlanErr).To(MatchError(errors.New("it broke")))
				})

				It("returns all warnings", func() {
					Expect(enablePlanWarnings).To(ConsistOf([]string{"foo", "bar"}))
				})
			})

			When("getting service plans fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicesReturns(
						[]ccv2.Service{
							{Label: "service-1", GUID: "service-guid-1"},
						},
						[]string{"foo"},
						nil)

					fakeCloudControllerClient.GetServicePlansReturns(nil, []string{"bar"}, errors.New("it broke"))
				})

				It("returns the error", func() {
					Expect(enablePlanErr).To(MatchError(errors.New("it broke")))
				})

				It("returns all warnings", func() {
					Expect(enablePlanWarnings).To(ConsistOf([]string{"foo", "bar"}))
				})
			})

			When("create service plan visibility fails", func() {
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

					fakeCloudControllerClient.CreateServicePlanVisibilityReturns(ccv2.ServicePlanVisibility{}, []string{"baz"}, errors.New("some error"))
				})

				It("returns all warnings", func() {
					Expect(enablePlanWarnings).To(ConsistOf([]string{"foo", "bar", "baz"}))
				})

				It("returns the error", func() {
					Expect(enablePlanErr).To(MatchError(errors.New("some error")))
				})
			})
		})

		When("the specified org does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv2.Organization{}, nil, nil)
			})

			It("returns an organization not found error", func() {
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

			It("returns service not found error", func() {
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

			It("returns an organization not found error", func() {
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
					nil,
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
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{},
					ccv2.Warnings{"foo", "bar"},
					errors.New("this is very bad"))
			})

			It("returns the error", func() {
				Expect(enableServiceForOrgErr).To(MatchError(errors.New("this is very bad")))
			})

			It("returns all warnings", func() {
				Expect(enableServiceForOrgWarnings).To(ConsistOf("foo", "bar"))
			})
		})

		When("getting organizations fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{{Label: "service-1", GUID: "service-guid-1"}},
					ccv2.Warnings{"foo"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{Name: "my-org", GUID: "org-guid-1"}},
					ccv2.Warnings{"bar"},
					errors.New("this is very bad"))
			})

			It("returns the error", func() {
				Expect(enableServiceForOrgErr).To(MatchError(errors.New("this is very bad")))
			})

			It("returns all warnings", func() {
				Expect(enableServiceForOrgWarnings).To(ConsistOf("foo", "bar"))
			})
		})

		When("getting service plans fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{{Label: "service-1", GUID: "service-guid-1"}},
					ccv2.Warnings{"foo"},
					nil)
				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv2.ServicePlan{},
					ccv2.Warnings{"bar"},
					errors.New("this is very bad"))
			})

			It("returns the error", func() {
				Expect(enableServiceForOrgErr).To(MatchError(errors.New("this is very bad")))
			})

			It("returns all warnings", func() {
				Expect(enableServiceForOrgWarnings).To(ConsistOf("foo", "bar"))
			})
		})

		When("creating service plan visibility fails", func() {
			BeforeEach(func() {
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
					errors.New("this is very bad"))
			})

			It("returns the error", func() {
				Expect(enableServiceForOrgErr).To(MatchError(errors.New("this is very bad")))
			})

			It("returns all warnings", func() {
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
				nil, nil)

			fakeCloudControllerClient.GetServicePlansReturns(
				[]ccv2.ServicePlan{
					{Name: "plan-1", GUID: "service-plan-guid-1"},
					{Name: "plan-2", GUID: "service-plan-guid-2"},
				},
				nil, nil)
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

			It("returns service not found error", func() {
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
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(nil, []string{"foo", "bar"}, errors.New("it broke"))
			})

			It("returns the error", func() {
				Expect(enableServiceErr).To(MatchError(errors.New("it broke")))
			})

			It("returns all warnings", func() {
				Expect(enableServiceWarnings).To(ConsistOf([]string{"foo", "bar"}))
			})
		})

		When("getting service plans fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{
						{Label: "service-1", GUID: "service-guid-1"},
					},
					[]string{"foo", "bar"},
					nil)

				fakeCloudControllerClient.GetServicePlansReturns(nil, []string{"baz"}, errors.New("it broke"))
			})

			It("returns all warnings", func() {
				Expect(enableServiceWarnings).To(ConsistOf([]string{"foo", "bar", "baz"}))
			})

			It("returns the error", func() {
				Expect(enableServiceErr).To(MatchError(errors.New("it broke")))
			})
		})

		When("update service plan fails", func() {
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

				fakeCloudControllerClient.UpdateServicePlanReturns([]string{"baz", "quux"}, errors.New("some error"))
			})

			It("returns all warnings", func() {
				Expect(enableServiceWarnings).To(ConsistOf([]string{"foo", "bar", "baz", "quux"}))
			})

			It("returns the error", func() {
				Expect(enableServiceErr).To(MatchError(errors.New("some error")))
			})
		})
	})

	Describe("DisableServiceForAllOrgs", func() {
		var (
			disableServiceErr      error
			disableServiceWarnings Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServicesReturns(
				[]ccv2.Service{
					{Label: "service-1", GUID: "service-guid-1"},
				},
				nil, nil)

			fakeCloudControllerClient.GetServicePlansReturns(
				[]ccv2.ServicePlan{
					{Name: "plan-1", GUID: "service-plan-guid-1", Public: true},
					{Name: "plan-2", GUID: "service-plan-guid-2", Public: true},
					{Name: "plan-3", GUID: "service-plan-guid-3", Public: false},
				},
				nil, nil)
		})

		JustBeforeEach(func() {
			disableServiceWarnings, disableServiceErr = actor.DisableServiceForAllOrgs("service-1")
		})

		It("should update all public plans to non-public", func() {
			Expect(disableServiceErr).NotTo(HaveOccurred())
			Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.UpdateServicePlanCallCount()).To(Equal(2))
			planGuid, public := fakeCloudControllerClient.UpdateServicePlanArgsForCall(0)
			Expect(planGuid).To(Equal("service-plan-guid-1"))
			Expect(public).To(BeFalse())
			planGuid, public = fakeCloudControllerClient.UpdateServicePlanArgsForCall(1)
			Expect(planGuid).To(Equal("service-plan-guid-2"))
			Expect(public).To(BeFalse())
		})

		When("the plan is enabled for an org", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlanVisibilitiesReturnsOnCall(2,
					[]ccv2.ServicePlanVisibility{{GUID: "service-plan-visibility-guid-1"}},
					nil, nil)
			})

			It("should delete the service plan visibility for the plan", func() {
				Expect(fakeCloudControllerClient.DeleteServicePlanVisibilityCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.DeleteServicePlanVisibilityArgsForCall(0)).To(Equal("service-plan-visibility-guid-1"))
			})
		})

		When("the service does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{},
					nil, nil)
			})

			It("returns service not found error", func() {
				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(disableServiceErr).To(MatchError(actionerror.ServiceNotFoundError{Name: "service-1"}))

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
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(nil, []string{"foo", "bar"}, errors.New("it broke"))
			})

			It("returns the error", func() {
				Expect(disableServiceErr).To(MatchError(errors.New("it broke")))
			})

			It("returns all warnings", func() {
				Expect(disableServiceWarnings).To(ConsistOf([]string{"foo", "bar"}))
			})
		})

		When("getting service plans fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{
						{Label: "service-1", GUID: "service-guid-1"},
					},
					[]string{"foo", "bar"},
					nil)

				fakeCloudControllerClient.GetServicePlansReturns(nil, []string{"baz"}, errors.New("it broke"))
			})

			It("returns all warnings", func() {
				Expect(disableServiceWarnings).To(ConsistOf([]string{"foo", "bar", "baz"}))
			})

			It("returns the error", func() {
				Expect(disableServiceErr).To(MatchError(errors.New("it broke")))
			})
		})

		When("update service plan fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{
						{Label: "service-1", GUID: "service-guid-1"},
					},
					[]string{"foo"},
					nil)

				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv2.ServicePlan{
						{Name: "plan-1", GUID: "service-plan-guid-1", Public: true},
						{Name: "plan-2", GUID: "service-plan-guid-2", Public: true},
					},
					[]string{"bar"},
					nil)

				fakeCloudControllerClient.UpdateServicePlanReturns([]string{"baz", "quux"}, errors.New("some error"))
			})

			It("returns all warnings", func() {
				Expect(disableServiceWarnings).To(ConsistOf([]string{"foo", "bar", "baz", "quux"}))
			})

			It("returns the error", func() {
				Expect(disableServiceErr).To(MatchError(errors.New("some error")))
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
						{Name: "plan-1", GUID: "service-plan-guid-1", Public: true},
						{Name: "plan-2", GUID: "service-plan-guid-2", Public: true},
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

			It("returns the warnings", func() {
				Expect(disableServiceWarnings).To(ConsistOf([]string{"foo", "bar", "baz", "qux", "baz", "qux"}))
			})
		})
	})

	Describe("DisablePlanForAllOrgs", func() {
		var (
			disablePlanErr      error
			disablePlanWarnings Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetServicesReturns(
				[]ccv2.Service{
					{Label: "service-1", GUID: "service-guid-1"},
				},
				nil, nil)
		})

		JustBeforeEach(func() {
			disablePlanWarnings, disablePlanErr = actor.DisablePlanForAllOrgs("service-1", "plan-2")
		})

		When("the plan is public", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv2.ServicePlan{
						{Name: "plan-1", GUID: "service-plan-guid-1", Public: true},
						{Name: "plan-2", GUID: "service-plan-guid-2", Public: true},
					},
					nil, nil)
			})

			It("updates the service plan to be not public", func() {
				Expect(disablePlanErr).NotTo(HaveOccurred())
				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateServicePlanCallCount()).To(Equal(1))

				planGuid, visible := fakeCloudControllerClient.UpdateServicePlanArgsForCall(0)
				Expect(planGuid).To(Equal("service-plan-guid-2"))
				Expect(visible).To(BeFalse())
			})

			When("the plan is visible in some orgs", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlanVisibilitiesReturns(
						[]ccv2.ServicePlanVisibility{
							{GUID: "service-visibility-guid-1"},
							{GUID: "service-visibility-guid-2"},
						},
						nil, nil)
				})

				It("disables the plan in both orgs", func() {
					Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesCallCount()).To(Equal(1))

					filters := fakeCloudControllerClient.GetServicePlanVisibilitiesArgsForCall(0)
					Expect(filters[0].Type).To(Equal(constant.ServicePlanGUIDFilter))
					Expect(filters[0].Operator).To(Equal(constant.EqualOperator))
					Expect(filters[0].Values).To(Equal([]string{"service-plan-guid-2"}))

					Expect(fakeCloudControllerClient.DeleteServicePlanVisibilityCallCount()).To(Equal(2))
				})

				It("and updates the service plan visibility for all orgs", func() {
					Expect(disablePlanErr).NotTo(HaveOccurred())
					Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.UpdateServicePlanCallCount()).To(Equal(1))

					planGuid, visible := fakeCloudControllerClient.UpdateServicePlanArgsForCall(0)
					Expect(planGuid).To(Equal("service-plan-guid-2"))
					Expect(visible).To(BeFalse())
				})

				When("deleting service plan visibilities fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.DeleteServicePlanVisibilityReturns(
							[]string{"visibility-warning"}, errors.New("delete-visibility-error"))
					})

					It("propagates the error", func() {
						Expect(disablePlanErr).To(MatchError(errors.New("delete-visibility-error")))
					})

					It("returns the warnings", func() {
						Expect(disablePlanWarnings).To(Equal(Warnings{"visibility-warning"}))
					})
				})
			})
		})

		When("the plan is not public", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv2.ServicePlan{
						{Name: "plan-1", GUID: "service-plan-guid-1", Public: true},
						{Name: "plan-2", GUID: "service-plan-guid-2", Public: false},
					},
					nil, nil)
			})

			It("does not update the plan to be not public", func() {
				Expect(disablePlanErr).NotTo(HaveOccurred())
				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.UpdateServicePlanCallCount()).To(Equal(0))
			})
		})

		When("getting service plan visibilities fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv2.ServicePlan{
						{Name: "plan-1", GUID: "service-plan-guid-1", Public: true},
						{Name: "plan-2", GUID: "service-plan-guid-2", Public: true},
					},
					nil, nil)
				fakeCloudControllerClient.GetServicePlanVisibilitiesReturns(
					[]ccv2.ServicePlanVisibility{},
					[]string{"a-warning", "another-warning"},
					errors.New("oh no"))
			})

			It("returns the error", func() {
				Expect(disablePlanErr).To(MatchError(errors.New("oh no")))
			})

			It("returns the warnings", func() {
				Expect(disablePlanWarnings).To(Equal(Warnings{"a-warning", "another-warning"}))
			})
		})

		When("getting service plans fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns(nil,
					[]string{"a-warning", "another-warning"},
					errors.New("it didn't work!"))
			})

			It("returns the error", func() {
				Expect(disablePlanErr).To(MatchError(errors.New("it didn't work!")))
			})

			It("returns the warnings", func() {
				Expect(disablePlanWarnings).To(Equal(Warnings{"a-warning", "another-warning"}))
			})
		})

		When("getting services fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(nil,
					[]string{"a-warning", "another-warning"},
					errors.New("it didn't work!"))
			})

			It("returns the error", func() {
				Expect(disablePlanErr).To(MatchError(errors.New("it didn't work!")))
			})

			It("returns the warnings", func() {
				Expect(disablePlanWarnings).To(Equal(Warnings{"a-warning", "another-warning"}))
			})
		})

		When("there are no matching services", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{}, []string{"warning-1", "warning-2"}, nil)
			})

			It("returns not found error", func() {
				Expect(disablePlanErr).To(MatchError(actionerror.ServiceNotFoundError{Name: "service-1"}))
			})

			It("returns all warnings", func() {
				Expect(disablePlanWarnings).To(ConsistOf([]string{"warning-1", "warning-2"}))
			})
		})

		When("there are no matching plans", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns([]ccv2.ServicePlan{}, []string{"warning-1", "warning-2"}, nil)
			})

			It("returns not found error", func() {
				Expect(disablePlanErr).To(MatchError(actionerror.ServicePlanNotFoundError{PlanName: "plan-2", ServiceName: "service-1"}))
			})

			It("returns all warnings", func() {
				Expect(disablePlanWarnings).To(ConsistOf([]string{"warning-1", "warning-2"}))
			})
		})

		When("updating service plan fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv2.ServicePlan{
						{Name: "plan-1", GUID: "service-plan-guid-1", Public: true},
						{Name: "plan-2", GUID: "service-plan-guid-2", Public: true},
					},
					nil, nil)
				fakeCloudControllerClient.UpdateServicePlanReturns(
					[]string{"a-warning", "another-warning"},
					errors.New("it didn't work!"))
			})

			It("returns the error", func() {
				Expect(disablePlanErr).To(MatchError(errors.New("it didn't work!")))
			})

			It("returns the warnings", func() {
				Expect(disablePlanWarnings).To(Equal(Warnings{"a-warning", "another-warning"}))
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
						{Name: "plan-1", GUID: "service-plan-guid-1", Public: true},
						{Name: "plan-2", GUID: "service-plan-guid-2", Public: true},
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

			It("returns the warnings", func() {
				Expect(disablePlanWarnings).To(ConsistOf([]string{"foo", "bar", "baz", "qux"}))
			})
		})
	})

	Describe("DisableServiceForOrg", func() {
		var disableServiceForOrgErr error
		var disableServiceForOrgWarnings Warnings

		JustBeforeEach(func() {
			disableServiceForOrgWarnings, disableServiceForOrgErr = actor.DisableServiceForOrg("service-1", "my-org")
		})

		When("the service and org exist", func() {
			BeforeEach(func() {
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
						{Name: "service-plan-1", GUID: "service-plan-guid-1"},
						{Name: "service-plan-2", GUID: "service-plan-guid-2"},
					},
					ccv2.Warnings{"baz"},
					nil)
				fakeCloudControllerClient.GetServicePlanVisibilitiesReturnsOnCall(0,
					[]ccv2.ServicePlanVisibility{
						{GUID: "service-visibility-guid-1"},
					},
					ccv2.Warnings{"qux-1"}, nil)
				fakeCloudControllerClient.GetServicePlanVisibilitiesReturnsOnCall(1,
					[]ccv2.ServicePlanVisibility{
						{GUID: "service-visibility-guid-2"},
					},
					ccv2.Warnings{"qux-2"}, nil)
			})

			It("deletes the service plan visibility for all plans for org", func() {
				Expect(disableServiceForOrgErr).NotTo(HaveOccurred())
				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				orgFilters := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(len(orgFilters)).To(Equal(1))
				Expect(orgFilters[0]).To(Equal(ccv2.Filter{
					Type:     constant.NameFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"my-org"},
				}))
				Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesCallCount()).To(Equal(2))
				filters := fakeCloudControllerClient.GetServicePlanVisibilitiesArgsForCall(0)

				Expect(filters).To(ConsistOf(
					ccv2.Filter{Type: constant.ServicePlanGUIDFilter, Operator: constant.EqualOperator, Values: []string{"service-plan-guid-1"}},
					ccv2.Filter{Type: constant.OrganizationGUIDFilter, Operator: constant.EqualOperator, Values: []string{"org-guid-1"}},
				))
				Expect(fakeCloudControllerClient.DeleteServicePlanVisibilityCallCount()).To(Equal(2))
				Expect(fakeCloudControllerClient.DeleteServicePlanVisibilityArgsForCall(0)).To(Equal("service-visibility-guid-1"))
				Expect(fakeCloudControllerClient.DeleteServicePlanVisibilityArgsForCall(1)).To(Equal("service-visibility-guid-2"))
			})

			When("there are warnings", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteServicePlanVisibilityReturns(
						[]string{"quux"},
						nil)
				})

				It("returns the warnings", func() {
					Expect(disableServiceForOrgWarnings).To(ConsistOf([]string{"foo", "bar", "baz", "qux-1", "quux", "qux-2", "quux"}))
				})
			})

			When("deleting service plan visibility fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteServicePlanVisibilityReturns(
						ccv2.Warnings{"quux"},
						errors.New("this is very bad"))
				})

				It("returns the error", func() {
					Expect(disableServiceForOrgErr).To(MatchError(errors.New("this is very bad")))
				})

				It("returns all warnings", func() {
					Expect(disableServiceForOrgWarnings).To(ConsistOf("foo", "bar", "baz", "qux-1", "quux"))
				})
			})
		})

		When("the service does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{},
					nil, nil)
			})

			It("returns service not found error", func() {
				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(disableServiceForOrgErr).To(MatchError(actionerror.ServiceNotFoundError{Name: "service-1"}))

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

			It("returns an organization not found error", func() {
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				Expect(disableServiceForOrgErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "my-org"}))
				filters := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(len(filters)).To(Equal(1))
				Expect(filters[0]).To(Equal(ccv2.Filter{
					Type:     constant.NameFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"my-org"},
				}))
			})
		})

		When("getting services fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{},
					ccv2.Warnings{"foo", "bar"},
					errors.New("this is very bad"))
			})

			It("returns the error", func() {
				Expect(disableServiceForOrgErr).To(MatchError(errors.New("this is very bad")))
			})

			It("returns all warnings", func() {
				Expect(disableServiceForOrgWarnings).To(ConsistOf("foo", "bar"))
			})
		})

		When("getting organizations fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{{Label: "service-1", GUID: "service-guid-1"}},
					ccv2.Warnings{"foo"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{Name: "my-org", GUID: "org-guid-1"}},
					ccv2.Warnings{"bar"},
					errors.New("this is very bad"))
			})

			It("returns the error", func() {
				Expect(disableServiceForOrgErr).To(MatchError(errors.New("this is very bad")))
			})

			It("returns all warnings", func() {
				Expect(disableServiceForOrgWarnings).To(ConsistOf("foo", "bar"))
			})
		})

		When("getting service plans fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{{Label: "service-1", GUID: "service-guid-1"}},
					ccv2.Warnings{"foo"},
					nil)
				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv2.ServicePlan{},
					ccv2.Warnings{"bar"},
					errors.New("this is very bad"))
			})

			It("returns the error", func() {
				Expect(disableServiceForOrgErr).To(MatchError(errors.New("this is very bad")))
			})

			It("returns all warnings", func() {
				Expect(disableServiceForOrgWarnings).To(ConsistOf("foo", "bar"))
			})
		})
	})

	Describe("DisablePlanForOrg", func() {
		var disablePlanForOrgErr error
		var disablePlanForOrgWarnings Warnings

		JustBeforeEach(func() {
			disablePlanForOrgWarnings, disablePlanForOrgErr = actor.DisablePlanForOrg("service-1", "service-plan-1", "my-org")
		})

		When("the service and plan and org exist", func() {
			BeforeEach(func() {
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
						{Name: "service-plan-1", GUID: "service-plan-guid-1"},
						{Name: "service-plan-2", GUID: "service-plan-guid-2"},
					},
					ccv2.Warnings{"baz"},
					nil)
				fakeCloudControllerClient.GetServicePlanVisibilitiesReturnsOnCall(0,
					[]ccv2.ServicePlanVisibility{
						{GUID: "service-visibility-guid-1"},
					},
					nil, nil)
			})

			It("deletes the service plan visibility for all plans for org", func() {
				Expect(disablePlanForOrgErr).NotTo(HaveOccurred())
				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				orgFilters := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(len(orgFilters)).To(Equal(1))
				Expect(orgFilters[0]).To(Equal(ccv2.Filter{
					Type:     constant.NameFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"my-org"},
				}))
				Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesCallCount()).To(Equal(1))
				filters := fakeCloudControllerClient.GetServicePlanVisibilitiesArgsForCall(0)

				Expect(filters).To(ConsistOf(
					ccv2.Filter{Type: constant.ServicePlanGUIDFilter, Operator: constant.EqualOperator, Values: []string{"service-plan-guid-1"}},
					ccv2.Filter{Type: constant.OrganizationGUIDFilter, Operator: constant.EqualOperator, Values: []string{"org-guid-1"}},
				))
				Expect(fakeCloudControllerClient.DeleteServicePlanVisibilityCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.DeleteServicePlanVisibilityArgsForCall(0)).To(Equal("service-visibility-guid-1"))
			})

			When("deleting service plan visibility fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteServicePlanVisibilityReturns(
						ccv2.Warnings{"qux"},
						errors.New("this is very bad"))
				})

				It("returns the error", func() {
					Expect(disablePlanForOrgErr).To(MatchError(errors.New("this is very bad")))
				})

				It("returns all warnings", func() {
					Expect(disablePlanForOrgWarnings).To(ConsistOf("foo", "bar", "baz", "qux"))
				})
			})
		})

		When("the service does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{},
					nil, nil)
			})

			It("returns service not found error", func() {
				Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
				Expect(disablePlanForOrgErr).To(MatchError(actionerror.ServiceNotFoundError{Name: "service-1"}))

				filters := fakeCloudControllerClient.GetServicesArgsForCall(0)
				Expect(len(filters)).To(Equal(1))
				Expect(filters[0]).To(Equal(ccv2.Filter{
					Type:     constant.LabelFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"service-1"},
				}))
			})
		})

		When("the service plan does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{{Label: "service-1", GUID: "service-guid-1"}},
					ccv2.Warnings{"foo"},
					nil)
				fakeCloudControllerClient.GetServicePlansReturns([]ccv2.ServicePlan{}, []string{"warning-1", "warning-2"}, nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{Name: "my-org", GUID: "org-guid-1"}},
					ccv2.Warnings{"bar"},
					nil)
			})

			It("returns not found error", func() {
				Expect(disablePlanForOrgErr).To(MatchError(actionerror.ServicePlanNotFoundError{PlanName: "service-plan-1", ServiceName: "service-1"}))
			})

			It("returns all warnings", func() {
				Expect(disablePlanForOrgWarnings).To(ConsistOf([]string{"foo", "warning-1", "warning-2", "bar"}))
			})
		})

		When("the specified org does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{{Label: "service-1", GUID: "service-guid-1"}},
					nil, nil)
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv2.Organization{}, nil, nil)
			})

			It("returns an organization not found error", func() {
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				Expect(disablePlanForOrgErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "my-org"}))
				filters := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(len(filters)).To(Equal(1))
				Expect(filters[0]).To(Equal(ccv2.Filter{
					Type:     constant.NameFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"my-org"},
				}))
			})
		})

		When("getting services fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{},
					ccv2.Warnings{"foo", "bar"},
					errors.New("this is very bad"))
			})

			It("returns the error", func() {
				Expect(disablePlanForOrgErr).To(MatchError(errors.New("this is very bad")))
			})

			It("returns all warnings", func() {
				Expect(disablePlanForOrgWarnings).To(ConsistOf("foo", "bar"))
			})
		})

		When("getting organizations fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{{Label: "service-1", GUID: "service-guid-1"}},
					ccv2.Warnings{"foo"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{Name: "my-org", GUID: "org-guid-1"}},
					ccv2.Warnings{"bar"},
					errors.New("this is very bad"))
			})

			It("returns the error", func() {
				Expect(disablePlanForOrgErr).To(MatchError(errors.New("this is very bad")))
			})

			It("returns all warnings", func() {
				Expect(disablePlanForOrgWarnings).To(ConsistOf("foo", "bar"))
			})
		})

		When("getting service plans fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{{Label: "service-1", GUID: "service-guid-1"}},
					ccv2.Warnings{"foo"},
					nil)
				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv2.ServicePlan{},
					ccv2.Warnings{"bar"},
					errors.New("this is very bad"))
			})

			It("returns the error", func() {
				Expect(disablePlanForOrgErr).To(MatchError(errors.New("this is very bad")))
			})

			It("returns all warnings", func() {
				Expect(disablePlanForOrgWarnings).To(ConsistOf("foo", "bar"))
			})
		})
	})
})
