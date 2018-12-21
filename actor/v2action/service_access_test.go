package v2action_test

import (
	"errors"
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

			It("first disables the plan in both orgs", func() {
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
					fmt.Errorf("oh no"))
			})

			It("returns the error", func() {
				Expect(enablePlanErr).To(MatchError(fmt.Errorf("oh no")))
			})

			It("returns the warnings", func() {
				Expect(enablePlanWarnings).To(Equal(Warnings{"a-warning", "another-warning"}))
			})
		})

		When("getting service plans fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansReturns(nil,
					[]string{"a-warning", "another-warning"},
					fmt.Errorf("it didn't work!"))
			})

			It("returns the error", func() {
				Expect(enablePlanErr).To(MatchError(fmt.Errorf("it didn't work!")))
			})

			It("returns the warnings", func() {
				Expect(enablePlanWarnings).To(Equal(Warnings{"a-warning", "another-warning"}))
			})
		})

		When("getting services fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(nil,
					[]string{"a-warning", "another-warning"},
					fmt.Errorf("it didn't work!"))
			})

			It("returns the error", func() {
				Expect(enablePlanErr).To(MatchError(fmt.Errorf("it didn't work!")))
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
					fmt.Errorf("it didn't work!"))
			})

			It("returns the error", func() {
				Expect(enablePlanErr).To(MatchError(fmt.Errorf("it didn't work!")))
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

			When("the plan is already globally enabled", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns(
						[]ccv2.ServicePlan{
							{Name: "plan-2", GUID: "service-plan-guid-2", Public: true},
						},
						nil, nil)
				})

				It("should not create org visibility", func() {
					Expect(enablePlanErr).ToNot(HaveOccurred())
					Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))

					Expect(fakeCloudControllerClient.CreateServicePlanVisibilityCallCount()).To(Equal(0))
				})
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
					fakeCloudControllerClient.GetServicesReturns(nil, []string{"foo", "bar"}, fmt.Errorf("it broke"))
				})

				It("returns the error", func() {
					Expect(enablePlanErr).To(MatchError(fmt.Errorf("it broke")))
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

					fakeCloudControllerClient.GetServicePlansReturns(nil, []string{"bar"}, fmt.Errorf("it broke"))
				})

				It("returns the error", func() {
					Expect(enablePlanErr).To(MatchError(fmt.Errorf("it broke")))
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

					fakeCloudControllerClient.CreateServicePlanVisibilityReturns(ccv2.ServicePlanVisibility{}, []string{"baz"}, fmt.Errorf("some error"))
				})

				It("returns all warnings", func() {
					Expect(enablePlanWarnings).To(ConsistOf([]string{"foo", "bar", "baz"}))
				})

				It("returns the error", func() {
					Expect(enablePlanErr).To(MatchError(fmt.Errorf("some error")))
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

			When("service is already enabled for all orgs for a plan", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns(
						[]ccv2.ServicePlan{
							{Name: "plan-1", GUID: "service-plan-guid-1", Public: true},
							{Name: "plan-2", GUID: "service-plan-guid-2", Public: false},
						},
						nil,
						nil)
				})

				It("does not create service plan visibility for the org", func() {
					Expect(enableServiceForOrgErr).NotTo(HaveOccurred())
					Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))

					Expect(fakeCloudControllerClient.CreateServicePlanVisibilityCallCount()).To(Equal(1))
					planGUID1, orgGUID1 := fakeCloudControllerClient.CreateServicePlanVisibilityArgsForCall(0)
					Expect(orgGUID1).To(Equal("org-guid-1"))
					Expect(planGUID1).To(Equal("service-plan-guid-2"))
				})
			})
		})

		When("getting services fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns(
					[]ccv2.Service{},
					ccv2.Warnings{"foo", "bar"},
					fmt.Errorf("this is very bad"))
			})

			It("returns the error", func() {
				Expect(enableServiceForOrgErr).To(MatchError(fmt.Errorf("this is very bad")))
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
					fmt.Errorf("this is very bad"))
			})

			It("returns the error", func() {
				Expect(enableServiceForOrgErr).To(MatchError(fmt.Errorf("this is very bad")))
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
					fmt.Errorf("this is very bad"))
			})

			It("returns the error", func() {
				Expect(enableServiceForOrgErr).To(MatchError(fmt.Errorf("this is very bad")))
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
					fmt.Errorf("this is very bad"))
			})

			It("returns the error", func() {
				Expect(enableServiceForOrgErr).To(MatchError(fmt.Errorf("this is very bad")))
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

		When("the service is already enabled for orgs individually", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlanVisibilitiesReturns(
					[]ccv2.ServicePlanVisibility{
						{GUID: "service-visibility-guid-1"},
					},
					nil, nil)
			})

			It("first disables the plan in both orgs", func() {
				Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesCallCount()).To(Equal(2))

				filters := fakeCloudControllerClient.GetServicePlanVisibilitiesArgsForCall(0)
				Expect(filters[0].Type).To(Equal(constant.ServicePlanGUIDFilter))
				Expect(filters[0].Operator).To(Equal(constant.EqualOperator))
				Expect(filters[0].Values).To(Equal([]string{"service-plan-guid-1"}))

				Expect(fakeCloudControllerClient.DeleteServicePlanVisibilityCallCount()).To(Equal(2))
			})

			When("deleting service plan visibilities fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteServicePlanVisibilityReturns(
						[]string{"visibility-warning"}, errors.New("delete-visibility-error"))
				})

				It("propagates the error", func() {
					Expect(enableServiceErr).To(MatchError(errors.New("delete-visibility-error")))
				})

				It("returns the warnings", func() {
					Expect(enableServiceWarnings).To(Equal(Warnings{"visibility-warning"}))
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
				fakeCloudControllerClient.GetServicesReturns(nil, []string{"foo", "bar"}, fmt.Errorf("it broke"))
			})

			It("returns the error", func() {
				Expect(enableServiceErr).To(MatchError(fmt.Errorf("it broke")))
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

				fakeCloudControllerClient.GetServicePlansReturns(nil, []string{"baz"}, fmt.Errorf("it broke"))
			})

			It("returns all warnings", func() {
				Expect(enableServiceWarnings).To(ConsistOf([]string{"foo", "bar", "baz"}))
			})

			It("returns the error", func() {
				Expect(enableServiceErr).To(MatchError(fmt.Errorf("it broke")))
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

				fakeCloudControllerClient.UpdateServicePlanReturns([]string{"baz", "quux"}, fmt.Errorf("some error"))
			})

			It("returns all warnings", func() {
				Expect(enableServiceWarnings).To(ConsistOf([]string{"foo", "bar", "baz", "quux"}))
			})

			It("returns the error", func() {
				Expect(enableServiceErr).To(MatchError(fmt.Errorf("some error")))
			})
		})
	})
})
