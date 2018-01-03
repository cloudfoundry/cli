package v2action_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Security Group Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetSecurityGroupsWithOrganizationSpaceAndLifecycle", func() {
		var (
			secGroupOrgSpaces []SecurityGroupWithOrganizationSpaceAndLifecycle
			includeStaging    bool
			warnings          Warnings
			err               error
		)

		JustBeforeEach(func() {
			secGroupOrgSpaces, warnings, err = actor.GetSecurityGroupsWithOrganizationSpaceAndLifecycle(includeStaging)
		})

		Context("when an error occurs getting security groups", func() {
			var returnedError error

			BeforeEach(func() {
				returnedError = errors.New("get-security-groups-error")
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					nil,
					ccv2.Warnings{"warning-1", "warning-2"},
					returnedError,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(returnedError))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		Context("when an error occurs getting running spaces", func() {
			var returnedError error

			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{
						{
							GUID: "security-group-guid-1",
							Name: "security-group-1",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			Context("when the error is generic", func() {
				BeforeEach(func() {
					returnedError = errors.New("get-spaces-error")
					fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturns(
						nil,
						ccv2.Warnings{"warning-3", "warning-4"},
						returnedError,
					)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError(returnedError))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(BeNil())
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(0)).To(Equal("security-group-guid-1"))
				})
			})

			Context("when the error is a resource not found error", func() {
				BeforeEach(func() {
					returnedError = ccerror.ResourceNotFoundError{Message: "could not find security group"}
					fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturns(
						nil,
						ccv2.Warnings{"warning-3", "warning-4"},
						returnedError,
					)
				})

				It("returns all warnings and continues", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4", returnedError.Error()))
				})
			})
		})

		Context("when an error occurs getting staging spaces", func() {
			var returnedError error

			BeforeEach(func() {
				includeStaging = true

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{
						{
							GUID: "security-group-guid-1",
							Name: "security-group-1",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)

				fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturns(
					nil,
					ccv2.Warnings{"warning-3", "warning-4"},
					nil,
				)
			})

			Context("when the error is generic", func() {
				BeforeEach(func() {
					returnedError = errors.New("get-staging-spaces-error")
					fakeCloudControllerClient.GetStagingSpacesBySecurityGroupReturns(
						nil,
						ccv2.Warnings{"warning-5", "warning-6"},
						returnedError,
					)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError(returnedError))
					Expect(warnings).To(ConsistOf(
						"warning-1",
						"warning-2",
						"warning-3",
						"warning-4",
						"warning-5",
						"warning-6",
					))
					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(BeNil())
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(0)).To(Equal("security-group-guid-1"))
					Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupArgsForCall(0)).To(Equal("security-group-guid-1"))
				})
			})

			Context("when the error is a resource not found error", func() {
				BeforeEach(func() {
					returnedError = ccerror.ResourceNotFoundError{Message: "could not find security group"}
					fakeCloudControllerClient.GetStagingSpacesBySecurityGroupReturns(
						nil,
						ccv2.Warnings{"warning-5", "warning-6"},
						returnedError,
					)
				})

				It("returns all warnings and continues", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf(
						"warning-1",
						"warning-2",
						"warning-3",
						"warning-4",
						"warning-5",
						"warning-6",
						returnedError.Error()))
				})
			})
		})

		Context("when an error occurs getting an organization", func() {
			var returnedError error

			BeforeEach(func() {
				includeStaging = true

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{
						{
							GUID: "security-group-guid-1",
							Name: "security-group-1",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
				fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturns(
					[]ccv2.Space{
						{
							GUID:             "space-guid-11",
							Name:             "space-11",
							OrganizationGUID: "org-guid-11",
						},
					},
					ccv2.Warnings{"warning-3", "warning-4"},
					nil,
				)
				fakeCloudControllerClient.GetStagingSpacesBySecurityGroupReturns(
					[]ccv2.Space{
						{
							GUID:             "space-guid-12",
							Name:             "space-12",
							OrganizationGUID: "org-guid-12",
						},
					},
					ccv2.Warnings{"warning-5", "warning-6"},
					nil,
				)
			})

			Context("when the error is generic", func() {
				BeforeEach(func() {
					returnedError = errors.New("get-org-error")
					fakeCloudControllerClient.GetOrganizationReturns(
						ccv2.Organization{},
						ccv2.Warnings{"warning-7", "warning-8"},
						returnedError,
					)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError(returnedError))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4", "warning-5", "warning-6", "warning-7", "warning-8"))
					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(BeNil())
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(0)).To(Equal("security-group-guid-1"))
					Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupArgsForCall(0)).To(Equal("security-group-guid-1"))
					Expect(fakeCloudControllerClient.GetOrganizationCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(0)).To(Equal("org-guid-11"))
				})
			})

			Context("when the error is a resource not found error", func() {
				BeforeEach(func() {
					returnedError = ccerror.ResourceNotFoundError{Message: "could not find the org"}
					fakeCloudControllerClient.GetOrganizationReturnsOnCall(0,
						ccv2.Organization{},
						ccv2.Warnings{"warning-7", "warning-8"},
						returnedError,
					)
					fakeCloudControllerClient.GetOrganizationReturnsOnCall(1,
						ccv2.Organization{},
						ccv2.Warnings{"warning-9", "warning-10"},
						nil,
					)
				})

				It("returns all warnings and continues", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(BeEquivalentTo(Warnings{
						"warning-1", "warning-2",
						"warning-3", "warning-4",
						"warning-5", "warning-6",
						"warning-7", "warning-8",
						returnedError.Error(),
						"warning-9", "warning-10",
					}))
					Expect(fakeCloudControllerClient.GetOrganizationCallCount()).To(Equal(2))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(0)).To(Equal("org-guid-11"))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(1)).To(Equal("org-guid-12"))
				})
			})
		})

		Context("when no errors are encountered", func() {
			var (
				expectedSecurityGroup1 SecurityGroup
				expectedSecurityGroup2 SecurityGroup
				expectedSecurityGroup3 SecurityGroup
				expectedSecurityGroup4 SecurityGroup
				expectedSecurityGroup5 SecurityGroup
				expectedSecurityGroup6 SecurityGroup
				expectedSecurityGroup7 SecurityGroup

				expectedOrg11  Organization
				expectedOrg12  Organization
				expectedOrg13  Organization
				expectedOrg21  Organization
				expectedOrg23  Organization
				expectedOrg33  Organization
				expectedOrgAll Organization

				expectedSpace11  Space
				expectedSpace12  Space
				expectedSpace13  Space
				expectedSpace21  Space
				expectedSpace22  Space
				expectedSpace23  Space
				expectedSpace31  Space
				expectedSpace32  Space
				expectedSpace33  Space
				expectedSpaceAll Space
			)

			BeforeEach(func() {
				expectedSecurityGroup1 = SecurityGroup{
					GUID:           "security-group-guid-1",
					Name:           "security-group-1",
					RunningDefault: true,
				}
				expectedSecurityGroup2 = SecurityGroup{
					GUID:           "security-group-guid-2",
					Name:           "security-group-2",
					StagingDefault: true,
				}
				expectedSecurityGroup3 = SecurityGroup{
					GUID: "security-group-guid-3",
					Name: "security-group-3",
				}
				expectedSecurityGroup4 = SecurityGroup{
					GUID: "security-group-guid-4",
					Name: "security-group-4",
				}
				expectedSecurityGroup5 = SecurityGroup{
					GUID:           "security-group-guid-5",
					Name:           "security-group-5",
					RunningDefault: true,
				}
				expectedSecurityGroup6 = SecurityGroup{
					GUID:           "security-group-guid-6",
					Name:           "security-group-6",
					StagingDefault: true,
				}
				expectedSecurityGroup7 = SecurityGroup{
					GUID:           "security-group-guid-7",
					Name:           "security-group-7",
					RunningDefault: true,
					StagingDefault: true,
				}

				expectedOrg11 = Organization{
					GUID: "<<org-guid-11",
					Name: "org-11",
				}
				expectedOrg12 = Organization{
					GUID: "org-guid-12",
					Name: "org-12",
				}
				expectedOrg13 = Organization{
					GUID: "org-guid-13",
					Name: "org-13",
				}
				expectedOrg21 = Organization{
					GUID: "org-guid-21",
					Name: "org-21",
				}
				expectedOrg23 = Organization{
					GUID: "org-guid-23",
					Name: "org-23",
				}
				expectedOrg33 = Organization{
					GUID: "org-guid-33",
					Name: "org-33",
				}
				expectedOrgAll = Organization{
					Name: "",
				}

				expectedSpace11 = Space{
					GUID: "space-guid-11",
					Name: "space-11",
				}
				expectedSpace12 = Space{
					GUID: "space-guid-12",
					Name: "space-12",
				}
				expectedSpace13 = Space{
					GUID: "space-guid-13",
					Name: "space-13",
				}
				expectedSpace21 = Space{
					GUID: "space-guid-21",
					Name: "space-21",
				}
				expectedSpace22 = Space{
					GUID: "space-guid-22",
					Name: "space-22",
				}
				expectedSpace23 = Space{
					GUID: "space-guid-23",
					Name: "space-23",
				}
				expectedSpace31 = Space{
					GUID: "space-guid-31",
					Name: "space-31",
				}
				expectedSpace32 = Space{
					GUID: "space-guid-32",
					Name: "space-32",
				}
				expectedSpace33 = Space{
					GUID: "space-guid-33",
					Name: "space-33",
				}
				expectedSpaceAll = Space{
					Name: "",
				}

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{
						{
							GUID:           "security-group-guid-1",
							Name:           "security-group-1",
							RunningDefault: true,
						},
						{
							GUID:           "security-group-guid-2",
							Name:           "security-group-2",
							StagingDefault: true,
						},
						{
							GUID: "security-group-guid-3",
							Name: "security-group-3",
						},
						{
							GUID: "security-group-guid-4",
							Name: "security-group-4",
						},
						{
							GUID:           "security-group-guid-5",
							Name:           "security-group-5",
							RunningDefault: true,
						},
						{
							GUID:           "security-group-guid-6",
							Name:           "security-group-6",
							StagingDefault: true,
						},
						{
							GUID:           "security-group-guid-7",
							Name:           "security-group-7",
							RunningDefault: true,
							StagingDefault: true,
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)

				fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturnsOnCall(0,
					[]ccv2.Space{
						{
							GUID:             "space-guid-13",
							Name:             "space-13",
							OrganizationGUID: "org-guid-13",
						},
						{
							GUID:             "space-guid-12",
							Name:             "space-12",
							OrganizationGUID: "org-guid-12",
						},
						{
							GUID:             "space-guid-11",
							Name:             "space-11",
							OrganizationGUID: "<<org-guid-11",
						},
					},
					ccv2.Warnings{"warning-3", "warning-4"},
					nil,
				)

				fakeCloudControllerClient.GetStagingSpacesBySecurityGroupReturnsOnCall(0,
					[]ccv2.Space{
						{
							GUID:             "space-guid-13",
							Name:             "space-13",
							OrganizationGUID: "org-guid-13",
						},
						{
							GUID:             "space-guid-12",
							Name:             "space-12",
							OrganizationGUID: "org-guid-12",
						},
						{
							GUID:             "space-guid-11",
							Name:             "space-11",
							OrganizationGUID: "<<org-guid-11",
						},
					},
					ccv2.Warnings{"warning-5", "warning-6"},
					nil,
				)

				fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturnsOnCall(1,
					[]ccv2.Space{
						{
							GUID:             "space-guid-21",
							Name:             "space-21",
							OrganizationGUID: "org-guid-21",
						},
						{
							GUID:             "space-guid-23",
							Name:             "space-23",
							OrganizationGUID: "org-guid-23",
						},
						{
							GUID:             "space-guid-22",
							Name:             "space-22",
							OrganizationGUID: "<<org-guid-11",
						},
					},
					ccv2.Warnings{"warning-7", "warning-8"},
					nil,
				)
				fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturnsOnCall(2,
					[]ccv2.Space{},
					ccv2.Warnings{"warning-9", "warning-10"},
					nil,
				)
				fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturnsOnCall(3,
					[]ccv2.Space{
						{
							GUID:             "space-guid-31",
							Name:             "space-31",
							OrganizationGUID: "org-guid-23",
						},
						{
							GUID:             "space-guid-32",
							Name:             "space-32",
							OrganizationGUID: "<<org-guid-11",
						},
						{
							GUID:             "space-guid-33",
							Name:             "space-33",
							OrganizationGUID: "org-guid-33",
						},
					},
					ccv2.Warnings{"warning-11", "warning-12"},
					nil,
				)
				fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturnsOnCall(4,
					[]ccv2.Space{},
					ccv2.Warnings{"warning-31", "warning-32"},
					nil,
				)
				fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturnsOnCall(5,
					[]ccv2.Space{},
					ccv2.Warnings{"warning-33", "warning-34"},
					nil,
				)
				fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturnsOnCall(6,
					[]ccv2.Space{},
					ccv2.Warnings{"warning-35", "warning-36"},
					nil,
				)
				fakeCloudControllerClient.GetOrganizationReturnsOnCall(0,
					ccv2.Organization{
						GUID: "org-guid-13",
						Name: "org-13",
					},
					ccv2.Warnings{"warning-13", "warning-14"},
					nil,
				)
				fakeCloudControllerClient.GetOrganizationReturnsOnCall(1,
					ccv2.Organization{
						GUID: "org-guid-12",
						Name: "org-12",
					},
					ccv2.Warnings{"warning-15", "warning-16"},
					nil,
				)
				fakeCloudControllerClient.GetOrganizationReturnsOnCall(2,
					ccv2.Organization{
						GUID: "<<org-guid-11",
						Name: "org-11",
					},
					ccv2.Warnings{"warning-17", "warning-18"},
					nil,
				)
				fakeCloudControllerClient.GetOrganizationReturnsOnCall(3,
					ccv2.Organization{
						GUID: "org-guid-21",
						Name: "org-21",
					},
					ccv2.Warnings{"warning-19", "warning-20"},
					nil,
				)
				fakeCloudControllerClient.GetOrganizationReturnsOnCall(4,
					ccv2.Organization{
						GUID: "org-guid-23",
						Name: "org-23",
					},
					ccv2.Warnings{"warning-21", "warning-22"},
					nil,
				)
				fakeCloudControllerClient.GetOrganizationReturnsOnCall(5,
					ccv2.Organization{
						GUID: "org-guid-33",
						Name: "org-33",
					},
					ccv2.Warnings{"warning-25", "warning-26"},
					nil,
				)
			})

			Context("when security groups bound to spaces in the staging lifecycle are included", func() {
				BeforeEach(func() {
					includeStaging = true
				})

				It("returns a slice of SecurityGroupWithOrganizationSpaceAndLifecycle and all warnings", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf(
						"warning-1", "warning-2",
						"warning-3", "warning-4",
						"warning-5", "warning-6",
						"warning-7", "warning-8",
						"warning-9", "warning-10",
						"warning-11", "warning-12",
						"warning-13", "warning-14",
						"warning-15", "warning-16",
						"warning-17", "warning-18",
						"warning-19", "warning-20",
						"warning-21", "warning-22",
						"warning-25", "warning-26",
						"warning-31", "warning-32",
						"warning-33", "warning-34",
						"warning-35", "warning-36",
					))
					expected := []SecurityGroupWithOrganizationSpaceAndLifecycle{
						{
							SecurityGroup: &expectedSecurityGroup1,
							Organization:  &expectedOrgAll,
							Space:         &expectedSpaceAll,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup1,
							Organization:  &expectedOrg11,
							Space:         &expectedSpace11,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup1,
							Organization:  &expectedOrg11,
							Space:         &expectedSpace11,
							Lifecycle:     ccv2.SecurityGroupLifecycleStaging,
						},
						{
							SecurityGroup: &expectedSecurityGroup1,
							Organization:  &expectedOrg12,
							Space:         &expectedSpace12,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup1,
							Organization:  &expectedOrg12,
							Space:         &expectedSpace12,
							Lifecycle:     ccv2.SecurityGroupLifecycleStaging,
						},
						{
							SecurityGroup: &expectedSecurityGroup1,
							Organization:  &expectedOrg13,
							Space:         &expectedSpace13,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup1,
							Organization:  &expectedOrg13,
							Space:         &expectedSpace13,
							Lifecycle:     ccv2.SecurityGroupLifecycleStaging,
						},
						{
							SecurityGroup: &expectedSecurityGroup2,
							Organization:  &expectedOrgAll,
							Space:         &expectedSpaceAll,
							Lifecycle:     ccv2.SecurityGroupLifecycleStaging,
						},
						{
							SecurityGroup: &expectedSecurityGroup2,
							Organization:  &expectedOrg11,
							Space:         &expectedSpace22,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup2,
							Organization:  &expectedOrg21,
							Space:         &expectedSpace21,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup2,
							Organization:  &expectedOrg23,
							Space:         &expectedSpace23,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup3,
							Organization:  &Organization{},
							Space:         &Space{},
						},
						{
							SecurityGroup: &expectedSecurityGroup4,
							Organization:  &expectedOrg11,
							Space:         &expectedSpace32,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup4,
							Organization:  &expectedOrg23,
							Space:         &expectedSpace31,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup4,
							Organization:  &expectedOrg33,
							Space:         &expectedSpace33,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup5,
							Organization:  &expectedOrgAll,
							Space:         &expectedSpaceAll,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup6,
							Organization:  &expectedOrgAll,
							Space:         &expectedSpaceAll,
							Lifecycle:     ccv2.SecurityGroupLifecycleStaging,
						},
						{
							SecurityGroup: &expectedSecurityGroup7,
							Organization:  &expectedOrgAll,
							Space:         &expectedSpaceAll,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup7,
							Organization:  &expectedOrgAll,
							Space:         &expectedSpaceAll,
							Lifecycle:     ccv2.SecurityGroupLifecycleStaging,
						},
					}
					Expect(secGroupOrgSpaces).To(Equal(expected))
					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(BeNil())

					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupCallCount()).To(Equal(7))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(0)).To(Equal("security-group-guid-1"))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(1)).To(Equal("security-group-guid-2"))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(2)).To(Equal("security-group-guid-3"))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(3)).To(Equal("security-group-guid-4"))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(4)).To(Equal("security-group-guid-5"))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(5)).To(Equal("security-group-guid-6"))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(6)).To(Equal("security-group-guid-7"))

					Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupCallCount()).To(Equal(7))
					Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupArgsForCall(0)).To(Equal("security-group-guid-1"))
					Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupArgsForCall(1)).To(Equal("security-group-guid-2"))
					Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupArgsForCall(2)).To(Equal("security-group-guid-3"))
					Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupArgsForCall(3)).To(Equal("security-group-guid-4"))
					Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupArgsForCall(4)).To(Equal("security-group-guid-5"))
					Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupArgsForCall(5)).To(Equal("security-group-guid-6"))
					Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupArgsForCall(6)).To(Equal("security-group-guid-7"))

					Expect(fakeCloudControllerClient.GetOrganizationCallCount()).To(Equal(6))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(0)).To(Equal("org-guid-13"))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(1)).To(Equal("org-guid-12"))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(2)).To(Equal("<<org-guid-11"))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(3)).To(Equal("org-guid-21"))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(4)).To(Equal("org-guid-23"))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(5)).To(Equal("org-guid-33"))
				})
			})

			Context("when security groups bound to spaces in the staging lifecycle phase are not included", func() {
				BeforeEach(func() {
					includeStaging = false
				})

				It("returns a slice of SecurityGroupWithOrganizationSpaceAndLifecycle, excluding any which are bound only in the staging lifecycle phase, and all warnings", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf(
						"warning-1", "warning-2",
						"warning-3", "warning-4",
						"warning-7", "warning-8",
						"warning-9", "warning-10",
						"warning-11", "warning-12",
						"warning-13", "warning-14",
						"warning-15", "warning-16",
						"warning-17", "warning-18",
						"warning-19", "warning-20",
						"warning-21", "warning-22",
						"warning-25", "warning-26",
						"warning-31", "warning-32",
						"warning-33", "warning-34",
						"warning-35", "warning-36",
					))

					expected := []SecurityGroupWithOrganizationSpaceAndLifecycle{
						{
							SecurityGroup: &expectedSecurityGroup1,
							Organization:  &expectedOrgAll,
							Space:         &expectedSpaceAll,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup1,
							Organization:  &expectedOrg11,
							Space:         &expectedSpace11,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup1,
							Organization:  &expectedOrg12,
							Space:         &expectedSpace12,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup1,
							Organization:  &expectedOrg13,
							Space:         &expectedSpace13,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup2,
							Organization:  &expectedOrgAll,
							Space:         &expectedSpaceAll,
							Lifecycle:     ccv2.SecurityGroupLifecycleStaging,
						},
						{
							SecurityGroup: &expectedSecurityGroup2,
							Organization:  &expectedOrg11,
							Space:         &expectedSpace22,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup2,
							Organization:  &expectedOrg21,
							Space:         &expectedSpace21,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup2,
							Organization:  &expectedOrg23,
							Space:         &expectedSpace23,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup3,
							Organization:  &Organization{},
							Space:         &Space{},
						},
						{
							SecurityGroup: &expectedSecurityGroup4,
							Organization:  &expectedOrg11,
							Space:         &expectedSpace32,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup4,
							Organization:  &expectedOrg23,
							Space:         &expectedSpace31,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup4,
							Organization:  &expectedOrg33,
							Space:         &expectedSpace33,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup5,
							Organization:  &expectedOrgAll,
							Space:         &expectedSpaceAll,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup6,
							Organization:  &expectedOrgAll,
							Space:         &expectedSpaceAll,
							Lifecycle:     ccv2.SecurityGroupLifecycleStaging,
						},
						{
							SecurityGroup: &expectedSecurityGroup7,
							Organization:  &expectedOrgAll,
							Space:         &expectedSpaceAll,
							Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
						},
						{
							SecurityGroup: &expectedSecurityGroup7,
							Organization:  &expectedOrgAll,
							Space:         &expectedSpaceAll,
							Lifecycle:     ccv2.SecurityGroupLifecycleStaging,
						},
					}
					Expect(secGroupOrgSpaces).To(Equal(expected))
					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(BeNil())

					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupCallCount()).To(Equal(7))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(0)).To(Equal("security-group-guid-1"))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(1)).To(Equal("security-group-guid-2"))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(2)).To(Equal("security-group-guid-3"))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(3)).To(Equal("security-group-guid-4"))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(4)).To(Equal("security-group-guid-5"))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(5)).To(Equal("security-group-guid-6"))
					Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(6)).To(Equal("security-group-guid-7"))

					Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupCallCount()).To(Equal(0))

					Expect(fakeCloudControllerClient.GetOrganizationCallCount()).To(Equal(6))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(0)).To(Equal("org-guid-13"))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(1)).To(Equal("org-guid-12"))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(2)).To(Equal("<<org-guid-11"))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(3)).To(Equal("org-guid-21"))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(4)).To(Equal("org-guid-23"))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(5)).To(Equal("org-guid-33"))
				})
			})
		})

		Context("when interleaved ResourceNotFoundErrors are encountered", func() {
			var (
				expectedSecurityGroup1 SecurityGroup
				expectedSecurityGroup2 SecurityGroup
				expectedSecurityGroup3 SecurityGroup

				expectedOrg13 Organization

				expectedSpace11 Space
				expectedSpace12 Space
			)

			BeforeEach(func() {
				includeStaging = true

				expectedSecurityGroup1 = SecurityGroup{
					GUID:           "security-group-guid-1",
					Name:           "security-group-1",
					RunningDefault: true,
				}
				expectedSecurityGroup2 = SecurityGroup{
					GUID:           "security-group-guid-2",
					Name:           "security-group-2",
					StagingDefault: true,
				}
				expectedSecurityGroup3 = SecurityGroup{
					GUID: "security-group-guid-3",
					Name: "security-group-3",
				}

				expectedOrg13 = Organization{
					GUID: "org-guid-13",
					Name: "org-13",
				}

				expectedSpace11 = Space{
					GUID: "space-guid-11",
					Name: "space-11",
				}
				expectedSpace12 = Space{
					GUID: "space-guid-12",
					Name: "space-12",
				}

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{
						ccv2.SecurityGroup(expectedSecurityGroup1),
						ccv2.SecurityGroup(expectedSecurityGroup2),
						ccv2.SecurityGroup(expectedSecurityGroup3),
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)

				fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturnsOnCall(0,
					nil,
					ccv2.Warnings{"warning-3", "warning-4"},
					ccerror.ResourceNotFoundError{Message: "running-error-1"},
				)

				fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturnsOnCall(1,
					[]ccv2.Space{
						{
							GUID:             "space-guid-13",
							Name:             "space-13",
							OrganizationGUID: "org-guid-13",
						},
						{
							GUID:             "space-guid-12",
							Name:             "space-12",
							OrganizationGUID: "org-guid-12",
						},
						{
							GUID:             "space-guid-11",
							Name:             "space-11",
							OrganizationGUID: "<<org-guid-11",
						},
					},
					ccv2.Warnings{"warning-5", "warning-6"},
					nil,
				)

				fakeCloudControllerClient.GetStagingSpacesBySecurityGroupReturnsOnCall(0,
					[]ccv2.Space{},
					ccv2.Warnings{"warning-7", "warning-8"},
					ccerror.ResourceNotFoundError{Message: "staging-error-1"},
				)

				fakeCloudControllerClient.GetRunningSpacesBySecurityGroupReturnsOnCall(2,
					[]ccv2.Space{
						{
							GUID:             "space-guid-11",
							Name:             "space-11",
							OrganizationGUID: "org-guid-13",
						},
					},
					ccv2.Warnings{"warning-9", "warning-10"},
					nil,
				)

				fakeCloudControllerClient.GetStagingSpacesBySecurityGroupReturnsOnCall(1,
					[]ccv2.Space{
						{
							GUID:             "space-guid-12",
							Name:             "space-12",
							OrganizationGUID: "org-guid-13",
						},
					},
					ccv2.Warnings{"warning-11", "warning-12"},
					nil,
				)

				fakeCloudControllerClient.GetOrganizationReturnsOnCall(0,
					ccv2.Organization(expectedOrg13),
					ccv2.Warnings{"warning-13", "warning-14"},
					nil,
				)
			})

			It("returns a slice of SecurityGroupWithOrganizationSpaceAndLifecycle, which have not errored-out, and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				// Used BeEquivalentTo instead of ConsistOf to enforce order.
				Expect(warnings).To(BeEquivalentTo([]string{
					"warning-1", "warning-2",
					"warning-3", "warning-4",
					"running-error-1",
					"warning-5", "warning-6",
					"warning-7", "warning-8",
					"staging-error-1",
					"warning-9", "warning-10",
					"warning-11", "warning-12",
					"warning-13", "warning-14",
				}))

				expected := []SecurityGroupWithOrganizationSpaceAndLifecycle{
					{
						SecurityGroup: &expectedSecurityGroup3,
						Organization:  &expectedOrg13,
						Space:         &expectedSpace11,
						Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
					},
					{
						SecurityGroup: &expectedSecurityGroup3,
						Organization:  &expectedOrg13,
						Space:         &expectedSpace12,
						Lifecycle:     ccv2.SecurityGroupLifecycleStaging,
					},
				}
				Expect(secGroupOrgSpaces).To(Equal(expected))
				Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(BeNil())

				Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupCallCount()).To(Equal(3))
				Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(0)).To(Equal("security-group-guid-1"))
				Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(1)).To(Equal("security-group-guid-2"))
				Expect(fakeCloudControllerClient.GetRunningSpacesBySecurityGroupArgsForCall(2)).To(Equal("security-group-guid-3"))

				Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupCallCount()).To(Equal(2))
				Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupArgsForCall(0)).To(Equal("security-group-guid-2"))
				Expect(fakeCloudControllerClient.GetStagingSpacesBySecurityGroupArgsForCall(1)).To(Equal("security-group-guid-3"))

				Expect(fakeCloudControllerClient.GetOrganizationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(0)).To(Equal("org-guid-13"))
			})
		})
	})

	Describe("GetSecurityGroupByName", func() {
		var (
			securityGroup SecurityGroup
			warnings      Warnings
			err           error
		)

		JustBeforeEach(func() {
			securityGroup, warnings, err = actor.GetSecurityGroupByName("some-security-group")
		})

		Context("when the security group exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{
						{
							GUID: "some-security-group-guid",
							Name: "some-security-group",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns the security group and all warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(securityGroup.GUID).To(Equal("some-security-group-guid"))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)
				Expect(query).To(Equal(
					[]ccv2.QQuery{{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-security-group"},
					}}))
			})
		})

		Context("when the security group does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns a SecurityGroupNotFound error", func() {
				Expect(err).To(MatchError(actionerror.SecurityGroupNotFoundError{Name: "some-security-group"}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		Context("an error occurs", func() {
			var returnedError error

			BeforeEach(func() {
				returnedError = errors.New("get-security-groups-error")
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{},
					ccv2.Warnings{"warning-1", "warning-2"},
					returnedError,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(returnedError))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("BindSecurityGroupToSpace", func() {
		var (
			lifecycle ccv2.SecurityGroupLifecycle
			err       error
			warnings  []string
		)

		JustBeforeEach(func() {
			warnings, err = actor.BindSecurityGroupToSpace("some-security-group-guid", "some-space-guid", lifecycle)
		})

		Context("when the lifecycle is neither running nor staging", func() {
			BeforeEach(func() {
				lifecycle = "bill & ted"
			})

			It("returns and appropriate error", func() {
				Expect(err).To(MatchError(fmt.Sprintf("Invalid lifecycle: %s", lifecycle)))
			})
		})

		Context("when the lifecycle is running", func() {
			BeforeEach(func() {
				lifecycle = ccv2.SecurityGroupLifecycleRunning
			})

			Context("when binding the space does not return an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.AssociateSpaceWithRunningSecurityGroupReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("returns warnings and no error", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(fakeCloudControllerClient.AssociateSpaceWithRunningSecurityGroupCallCount()).To(Equal(1))
					securityGroupGUID, spaceGUID := fakeCloudControllerClient.AssociateSpaceWithRunningSecurityGroupArgsForCall(0)
					Expect(securityGroupGUID).To(Equal("some-security-group-guid"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			Context("when binding the space returns an error", func() {
				var returnedError error
				BeforeEach(func() {
					returnedError = errors.New("associate-space-error")
					fakeCloudControllerClient.AssociateSpaceWithRunningSecurityGroupReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						returnedError,
					)
				})

				It("returns the error and warnings", func() {
					Expect(err).To(Equal(returnedError))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})

		Context("when the lifecycle is staging", func() {
			BeforeEach(func() {
				lifecycle = ccv2.SecurityGroupLifecycleStaging
			})

			Context("when binding the space does not return an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.AssociateSpaceWithStagingSecurityGroupReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("returns warnings and no error", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(fakeCloudControllerClient.AssociateSpaceWithStagingSecurityGroupCallCount()).To(Equal(1))
					securityGroupGUID, spaceGUID := fakeCloudControllerClient.AssociateSpaceWithStagingSecurityGroupArgsForCall(0)
					Expect(securityGroupGUID).To(Equal("some-security-group-guid"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			Context("when binding the space returns an error", func() {
				var returnedError error
				BeforeEach(func() {
					returnedError = errors.New("associate-space-error")
					fakeCloudControllerClient.AssociateSpaceWithStagingSecurityGroupReturns(
						ccv2.Warnings{"warning-1", "warning-2"},
						returnedError,
					)
				})

				It("returns the error and warnings", func() {
					Expect(err).To(Equal(returnedError))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})
	})

	Describe("GetSpaceRunningSecurityGroupsBySpace", func() {
		Context("when the space exists and there are no errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
					[]ccv2.SecurityGroup{
						{
							Name: "some-shared-security-group",
						},
						{
							Name: "some-running-security-group",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns the security groups and warnings", func() {
				securityGroups, warnings, err := actor.GetSpaceRunningSecurityGroupsBySpace("space-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf([]string{"warning-1", "warning-2"}))
				Expect(securityGroups).To(Equal(
					[]SecurityGroup{
						{
							Name: "some-shared-security-group",
						},
						{
							Name: "some-running-security-group",
						},
					}))

				Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
				spaceGUID, queries := fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)
				Expect(spaceGUID).To(Equal("space-guid"))
				Expect(queries).To(BeNil())
			})
		})

		Context("when the space does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
					nil,
					nil,
					ccerror.ResourceNotFoundError{})
			})

			It("returns an SpaceNotFoundError", func() {
				_, _, err := actor.GetSpaceRunningSecurityGroupsBySpace("space-guid")
				Expect(err).To(MatchError(actionerror.SpaceNotFoundError{GUID: "space-guid"}))
			})
		})

		Context("when there is an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("banana")
				fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
					nil,
					ccv2.Warnings{"warning-1", "warning-2"},
					expectedErr)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetSpaceRunningSecurityGroupsBySpace("space-guid")
				Expect(warnings).To(ConsistOf([]string{"warning-1", "warning-2"}))
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})

	Describe("GetSpaceStagingSecurityGroupsBySpace", func() {
		Context("when the space exists and there are no errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
					[]ccv2.SecurityGroup{
						{
							Name: "some-shared-security-group",
						},
						{
							Name: "some-staging-security-group",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns the security groups and warnings", func() {
				securityGroups, warnings, err := actor.GetSpaceStagingSecurityGroupsBySpace("space-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf([]string{"warning-1", "warning-2"}))
				Expect(securityGroups).To(Equal(
					[]SecurityGroup{
						{
							Name: "some-shared-security-group",
						},
						{
							Name: "some-staging-security-group",
						},
					}))

				Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(1))
				spaceGUID, queries := fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceArgsForCall(0)
				Expect(spaceGUID).To(Equal("space-guid"))
				Expect(queries).To(BeNil())
			})
		})

		Context("when the space does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
					nil,
					nil,
					ccerror.ResourceNotFoundError{})
			})

			It("returns an SpaceNotFoundError", func() {
				_, _, err := actor.GetSpaceStagingSecurityGroupsBySpace("space-guid")
				Expect(err).To(MatchError(actionerror.SpaceNotFoundError{GUID: "space-guid"}))
			})
		})

		Context("when there is an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("banana")
				fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
					nil,
					ccv2.Warnings{"warning-1", "warning-2"},
					expectedErr)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.GetSpaceStagingSecurityGroupsBySpace("space-guid")
				Expect(warnings).To(ConsistOf([]string{"warning-1", "warning-2"}))
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})

	Describe("UnbindSecurityGroupByNameAndSpace", func() {
		var (
			lifecycle ccv2.SecurityGroupLifecycle
			warnings  Warnings
			err       error
		)

		JustBeforeEach(func() {
			warnings, err = actor.UnbindSecurityGroupByNameAndSpace("some-security-group", "some-space-guid", lifecycle)
		})

		Context("when the requested lifecycle is neither running nor staging", func() {
			BeforeEach(func() {
				lifecycle = "bill & ted"
			})

			It("returns and appropriate error", func() {
				Expect(err).To(MatchError(fmt.Sprintf("Invalid lifecycle: %s", lifecycle)))
			})
		})

		Context("when the security group is not found", func() {
			BeforeEach(func() {
				lifecycle = ccv2.SecurityGroupLifecycleStaging

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{},
					ccv2.Warnings{"security-group-warning"},
					nil)
			})

			It("returns the error and all warnings", func() {
				Expect(warnings).To(ConsistOf([]string{"security-group-warning"}))
				Expect(err).To(MatchError(actionerror.SecurityGroupNotFoundError{Name: "some-security-group"}))
			})
		})

		Context("when an error is encountered fetching security groups", func() {
			var returnedError error

			BeforeEach(func() {
				lifecycle = ccv2.SecurityGroupLifecycleRunning

				returnedError = errors.New("get-security-groups-error")
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{},
					ccv2.Warnings{"warning-1", "warning-2"},
					returnedError)
			})

			It("returns all warnings", func() {
				Expect(err).To(MatchError(returnedError))
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})

		Context("when the requested lifecycle is running", func() {
			BeforeEach(func() {
				lifecycle = ccv2.SecurityGroupLifecycleRunning

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{{
						Name: "some-security-group",
						GUID: "some-security-group-guid",
					}},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
			})

			Context("when the security group is bound to running", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{
							{
								Name: "some-security-group",
								GUID: "some-security-group-guid",
							},
						},
						ccv2.Warnings{"warning-3", "warning-4"},
						nil,
					)
				})

				Context("when an error is encountered checking whether the security group is bound to the space in the running phase", func() {
					var returnedError error

					BeforeEach(func() {
						returnedError = errors.New("get-security-groups-error")
						fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
							[]ccv2.SecurityGroup{
								{
									Name: "some-security-group",
									GUID: "some-security-group-guid",
								},
							},
							ccv2.Warnings{"warning-3", "warning-4"},
							returnedError,
						)
					})

					It("returns all warnings", func() {
						Expect(err).To(MatchError(returnedError))
						Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2", "warning-3", "warning-4"}))
					})
				})

				Context("when an error is encountered unbinding the security group from the space", func() {
					var returnedError error

					BeforeEach(func() {
						returnedError = errors.New("associate-space-error")
						fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupReturns(
							ccv2.Warnings{"warning-5", "warning-6"},
							returnedError)
					})

					It("returns the error and all warnings", func() {
						Expect(warnings).To(ConsistOf([]string{
							"warning-1",
							"warning-2",
							"warning-3",
							"warning-4",
							"warning-5",
							"warning-6",
						}))
						Expect(err).To(MatchError(returnedError))
					})
				})

				Context("when no errors are encountered", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupReturns(
							ccv2.Warnings{"warning-5", "warning-6"},
							nil)
					})

					It("returns all warnings", func() {
						Expect(warnings).To(ConsistOf([]string{
							"warning-1",
							"warning-2",
							"warning-3",
							"warning-4",
							"warning-5",
							"warning-6",
						}))
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
						spaceGUID, queries := fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(queries).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(0))

						Expect(fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupCallCount()).To(Equal(1))
						securityGroupGUID, spaceGUID := fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupArgsForCall(0)
						Expect(securityGroupGUID).To(Equal("some-security-group-guid"))
						Expect(spaceGUID).To(Equal("some-space-guid"))

						Expect(fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupCallCount()).To(Equal(0))
					})
				})
			})

			Context("when the security group is bound to neither running nor staging", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-3", "warning-4"},
						nil,
					)
					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-5", "warning-6"},
						nil,
					)
				})

				Context("when no errors are encountered", func() {
					It("returns all warnings", func() {
						Expect(warnings).To(ConsistOf([]string{
							"warning-1",
							"warning-2",
							"warning-3",
							"warning-4",
							"warning-5",
							"warning-6",
						}))
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
						spaceGUIDRunning, queriesRunning := fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)
						Expect(spaceGUIDRunning).To(Equal("some-space-guid"))
						Expect(queriesRunning).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(1))
						spaceGUIDStaging, queriesStaging := fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceArgsForCall(0)
						Expect(spaceGUIDStaging).To(Equal("some-space-guid"))
						Expect(queriesStaging).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupCallCount()).To(Equal(0))
						Expect(fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupCallCount()).To(Equal(0))
					})
				})
			})

			Context("when the security group is bound to staging", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-3", "warning-4"},
						nil,
					)
					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{
							{
								Name: "some-security-group",
								GUID: "some-security-group-guid",
							},
						},
						ccv2.Warnings{"warning-5", "warning-6"},
						nil,
					)
				})

				It("returns all warnings and a SecurityGroupNotBoundError", func() {
					Expect(warnings).To(ConsistOf([]string{
						"warning-1",
						"warning-2",
						"warning-3",
						"warning-4",
						"warning-5",
						"warning-6",
					}))
					Expect(err).To(MatchError(actionerror.SecurityGroupNotBoundError{
						Name:      "some-security-group",
						Lifecycle: lifecycle,
					}))

					Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
					spaceGUIDRunning, queriesRunning := fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)
					Expect(spaceGUIDRunning).To(Equal("some-space-guid"))
					Expect(queriesRunning).To(Equal([]ccv2.QQuery{{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-security-group"},
					}}))

					Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(1))
					spaceGUIDStaging, queriesStaging := fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceArgsForCall(0)
					Expect(spaceGUIDStaging).To(Equal("some-space-guid"))
					Expect(queriesStaging).To(Equal([]ccv2.QQuery{{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-security-group"},
					}}))

					Expect(fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupCallCount()).To(Equal(0))
				})
			})
		})

		Context("when the requested lifecycle is staging", func() {
			BeforeEach(func() {
				lifecycle = ccv2.SecurityGroupLifecycleStaging

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{{
						Name: "some-security-group",
						GUID: "some-security-group-guid",
					}},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
			})

			Context("when an error is encountered checking whether the security group is bound to the space in the staging phase", func() {
				var returnedError error

				BeforeEach(func() {
					returnedError = errors.New("get-space-staging-security-groups-error")
					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-3", "warning-4"},
						returnedError,
					)
				})

				It("returns all warnings", func() {
					Expect(err).To(MatchError(returnedError))
					Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2", "warning-3", "warning-4"}))
				})
			})

			Context("when the security group is bound to staging", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{
							{
								Name: "some-security-group",
								GUID: "some-security-group-guid",
							},
						},
						ccv2.Warnings{"warning-3", "warning-4"},
						nil,
					)
				})

				Context("when an error is encountered unbinding the security group the space", func() {
					var returnedError error

					BeforeEach(func() {
						returnedError = errors.New("associate-space-error")
						fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupReturns(
							ccv2.Warnings{"warning-5", "warning-6"},
							returnedError)
					})

					It("returns the error and all warnings", func() {
						Expect(err).To(MatchError(returnedError))
						Expect(warnings).To(ConsistOf([]string{
							"warning-1",
							"warning-2",
							"warning-3",
							"warning-4",
							"warning-5",
							"warning-6",
						}))
					})
				})

				Context("when no errors are encountered", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupReturns(
							ccv2.Warnings{"warning-5", "warning-6"},
							nil)
					})

					It("unbinds and returns all warnings", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf([]string{
							"warning-1",
							"warning-2",
							"warning-3",
							"warning-4",
							"warning-5",
							"warning-6",
						}))

						Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(1))
						spaceGUID, queries := fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceArgsForCall(0)
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(queries).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(0))

						Expect(fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupCallCount()).To(Equal(1))
						securityGroupGUID, spaceGUID := fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupArgsForCall(0)
						Expect(securityGroupGUID).To(Equal("some-security-group-guid"))
						Expect(spaceGUID).To(Equal("some-space-guid"))

						Expect(fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupCallCount()).To(Equal(0))
					})
				})
			})

			Context("when the security group is bound to neither running nor staging", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-3", "warning-4"},
						nil,
					)
					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-5", "warning-6"},
						nil,
					)
				})

				Context("when no errors are encountered", func() {
					It("returns all warnings", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf([]string{
							"warning-1",
							"warning-2",
							"warning-3",
							"warning-4",
							"warning-5",
							"warning-6",
						}))

						Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(1))
						spaceGUIDStaging, queriesStaging := fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceArgsForCall(0)
						Expect(spaceGUIDStaging).To(Equal("some-space-guid"))
						Expect(queriesStaging).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
						spaceGUIDRunning, queriesRunning := fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)
						Expect(spaceGUIDRunning).To(Equal("some-space-guid"))
						Expect(queriesRunning).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupCallCount()).To(Equal(0))
						Expect(fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupCallCount()).To(Equal(0))
					})
				})
			})

			Context("when the security group is bound to running", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-3", "warning-4"},
						nil,
					)
					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{
							{
								Name: "some-security-group",
								GUID: "some-security-group-guid",
							},
						},
						ccv2.Warnings{"warning-5", "warning-6"},
						nil,
					)
				})

				It("returns all warnings and a SecurityGroupNotBoundError", func() {
					Expect(warnings).To(ConsistOf([]string{
						"warning-1",
						"warning-2",
						"warning-3",
						"warning-4",
						"warning-5",
						"warning-6",
					}))

					Expect(err).To(MatchError(actionerror.SecurityGroupNotBoundError{
						Name:      "some-security-group",
						Lifecycle: lifecycle,
					}))

					Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(1))
					spaceGUIDStaging, queriesStaging := fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceArgsForCall(0)
					Expect(spaceGUIDStaging).To(Equal("some-space-guid"))
					Expect(queriesStaging).To(Equal([]ccv2.QQuery{{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-security-group"},
					}}))

					Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
					spaceGUIDRunning, queriesRunning := fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)
					Expect(spaceGUIDRunning).To(Equal("some-space-guid"))
					Expect(queriesRunning).To(Equal([]ccv2.QQuery{{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-security-group"},
					}}))

					Expect(fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupCallCount()).To(Equal(0))
				})
			})

			Context("when it is not bound to staging and an error occurs checking whether bound to running", func() {
				var returnedError error

				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-3", "warning-4"},
						nil,
					)
					returnedError = errors.New("get-space-running-security-groups-error")
					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{
							{
								Name: "some-security-group",
								GUID: "some-security-group-guid",
							},
						},
						ccv2.Warnings{"warning-5", "warning-6"},
						returnedError,
					)
				})

				It("returns all warnings", func() {
					Expect(err).To(MatchError(returnedError))
					Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2", "warning-3", "warning-4", "warning-5", "warning-6"}))
				})
			})
		})
	})

	Describe("UnbindSecurityGroupByNameOrganizationNameAndSpaceName", func() {
		var (
			lifecycle ccv2.SecurityGroupLifecycle
			warnings  []string
			err       error
		)

		JustBeforeEach(func() {
			warnings, err = actor.UnbindSecurityGroupByNameOrganizationNameAndSpaceName("some-security-group", "some-org", "some-space", lifecycle)
		})

		Context("when the requested lifecycle is neither running nor staging", func() {
			BeforeEach(func() {
				lifecycle = "bill & ted"
			})

			It("returns and appropriate error", func() {
				Expect(err).To(MatchError(fmt.Sprintf("Invalid lifecycle: %s", lifecycle)))
			})
		})

		Context("when the security group is not found", func() {
			BeforeEach(func() {
				lifecycle = ccv2.SecurityGroupLifecycleRunning

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{},
					ccv2.Warnings{"security-group-warning"},
					nil)
			})

			It("returns the error and all warnings", func() {
				Expect(warnings).To(ConsistOf([]string{"security-group-warning"}))
				Expect(err).To(MatchError(actionerror.SecurityGroupNotFoundError{Name: "some-security-group"}))
			})
		})

		Context("when an error is encountered getting the organization", func() {
			BeforeEach(func() {
				lifecycle = ccv2.SecurityGroupLifecycleRunning

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{{
						Name: "some-security-group",
						GUID: "some-security-group-guid",
					}},
					ccv2.Warnings{"security-group-warning"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{"org-warning"},
					nil)
			})

			It("returns the error and all warnings", func() {
				Expect(warnings).To(ConsistOf([]string{"security-group-warning", "org-warning"}))
				Expect(err).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org"}))
			})
		})

		Context("when an error is encountered getting the space", func() {
			BeforeEach(func() {
				lifecycle = ccv2.SecurityGroupLifecycleRunning

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{{
						Name: "some-security-group",
						GUID: "some-security-group-guid",
					}},
					ccv2.Warnings{"security-group-warning"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{
						Name: "some-org",
						GUID: "some-org-guid",
					}},
					ccv2.Warnings{"org-warning"},
					nil)
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{},
					ccv2.Warnings{"space-warning"},
					nil)
			})

			It("returns the error and all warnings", func() {
				Expect(warnings).To(ConsistOf([]string{"security-group-warning", "org-warning", "space-warning"}))
				Expect(err).To(MatchError(actionerror.SpaceNotFoundError{Name: "some-space"}))
			})
		})

		Context("when the requested lifecycle is running", func() {
			BeforeEach(func() {
				lifecycle = ccv2.SecurityGroupLifecycleRunning

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{{
						Name: "some-security-group",
						GUID: "some-security-group-guid",
					}},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{
						Name: "some-org",
						GUID: "some-org-guid",
					}},
					ccv2.Warnings{"warning-3", "warning-4"},
					nil)
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{{
						Name: "some-space",
						GUID: "some-space-guid",
					}},
					ccv2.Warnings{"warning-5", "warning-6"},
					nil)
			})

			Context("when the security group is bound to running", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{
							{
								Name: "some-security-group",
								GUID: "some-security-group-guid",
							},
						},
						ccv2.Warnings{"warning-7", "warning-8"},
						nil,
					)
				})

				Context("when an error is encountered unbinding the security group from the space", func() {
					var returnedError error

					BeforeEach(func() {
						returnedError = errors.New("associate-space-error")
						fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupReturns(
							ccv2.Warnings{"warning-9", "warning-10"},
							returnedError)
					})

					It("returns the error and all warnings", func() {
						Expect(warnings).To(ConsistOf([]string{
							"warning-1",
							"warning-2",
							"warning-3",
							"warning-4",
							"warning-5",
							"warning-6",
							"warning-7",
							"warning-8",
							"warning-9",
							"warning-10",
						}))
						Expect(err).To(MatchError(returnedError))
					})
				})

				Context("when no errors are encountered", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupReturns(
							ccv2.Warnings{"warning-9", "warning-10"},
							nil)
					})

					It("returns all warnings", func() {
						Expect(warnings).To(ConsistOf([]string{
							"warning-1",
							"warning-2",
							"warning-3",
							"warning-4",
							"warning-5",
							"warning-6",
							"warning-7",
							"warning-8",
							"warning-9",
							"warning-10",
						}))
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-org"},
						}}))

						Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-space"},
						}, {
							Filter:   ccv2.OrganizationGUIDFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-org-guid"},
						}}))

						Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
						spaceGUID, queries := fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(queries).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(0))

						Expect(fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupCallCount()).To(Equal(1))
						securityGroupGUID, spaceGUID := fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupArgsForCall(0)
						Expect(securityGroupGUID).To(Equal("some-security-group-guid"))
						Expect(spaceGUID).To(Equal("some-space-guid"))

						Expect(fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupCallCount()).To(Equal(0))
					})
				})
			})

			Context("when the security group is bound to neither running nor staging", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-7", "warning-8"},
						nil,
					)
					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-9", "warning-10"},
						nil,
					)
				})

				Context("when no errors are encountered", func() {
					It("returns all warnings", func() {
						Expect(warnings).To(ConsistOf([]string{
							"warning-1",
							"warning-2",
							"warning-3",
							"warning-4",
							"warning-5",
							"warning-6",
							"warning-7",
							"warning-8",
							"warning-9",
							"warning-10",
						}))
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
						spaceGUIDRunning, queriesRunning := fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)
						Expect(spaceGUIDRunning).To(Equal("some-space-guid"))
						Expect(queriesRunning).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(1))
						spaceGUIDStaging, queriesStaging := fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceArgsForCall(0)
						Expect(spaceGUIDStaging).To(Equal("some-space-guid"))
						Expect(queriesStaging).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupCallCount()).To(Equal(0))
						Expect(fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupCallCount()).To(Equal(0))
					})
				})
			})

			Context("when the security group is bound to staging", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-7", "warning-8"},
						nil,
					)
					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{
							{
								Name: "some-security-group",
								GUID: "some-security-group-guid",
							},
						},
						ccv2.Warnings{"warning-9", "warning-10"},
						nil,
					)
				})

				It("returns all warnings and a SecurityGroupNotBoundError", func() {
					Expect(warnings).To(ConsistOf([]string{
						"warning-1",
						"warning-2",
						"warning-3",
						"warning-4",
						"warning-5",
						"warning-6",
						"warning-7",
						"warning-8",
						"warning-9",
						"warning-10",
					}))
					Expect(err).To(MatchError(actionerror.SecurityGroupNotBoundError{
						Name:      "some-security-group",
						Lifecycle: lifecycle,
					}))

					Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
					spaceGUIDRunning, queriesRunning := fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)
					Expect(spaceGUIDRunning).To(Equal("some-space-guid"))
					Expect(queriesRunning).To(Equal([]ccv2.QQuery{{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-security-group"},
					}}))

					Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(1))
					spaceGUIDStaging, queriesStaging := fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceArgsForCall(0)
					Expect(spaceGUIDStaging).To(Equal("some-space-guid"))
					Expect(queriesStaging).To(Equal([]ccv2.QQuery{{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-security-group"},
					}}))

					Expect(fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupCallCount()).To(Equal(0))
				})
			})
		})

		Context("when the requested lifecycle is staging", func() {
			BeforeEach(func() {
				lifecycle = ccv2.SecurityGroupLifecycleStaging

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]ccv2.SecurityGroup{{
						Name: "some-security-group",
						GUID: "some-security-group-guid",
					}},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{
						Name: "some-org",
						GUID: "some-org-guid",
					}},
					ccv2.Warnings{"warning-3", "warning-4"},
					nil)
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{{
						Name: "some-space",
						GUID: "some-space-guid",
					}},
					ccv2.Warnings{"warning-5", "warning-6"},
					nil)
			})

			Context("when the security group is bound to staging", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{
							{
								Name: "some-security-group",
								GUID: "some-security-group-guid",
							},
						},
						ccv2.Warnings{"warning-7", "warning-8"},
						nil,
					)
				})

				Context("when an error is encountered unbinding the security group the space", func() {
					var returnedError error

					BeforeEach(func() {
						fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
							[]ccv2.SecurityGroup{
								{
									Name: "some-security-group",
									GUID: "some-security-group-guid",
								},
							},
							ccv2.Warnings{"warning-7", "warning-8"},
							nil,
						)
						returnedError = errors.New("associate-space-error")
						fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupReturns(
							ccv2.Warnings{"warning-9", "warning-10"},
							returnedError)
					})

					It("returns the error and all warnings", func() {
						Expect(warnings).To(ConsistOf([]string{
							"warning-1",
							"warning-2",
							"warning-3",
							"warning-4",
							"warning-5",
							"warning-6",
							"warning-7",
							"warning-8",
							"warning-9",
							"warning-10",
						}))
						Expect(err).To(MatchError(returnedError))
					})
				})

				Context("when no errors are encountered", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupReturns(
							ccv2.Warnings{"warning-9", "warning-10"},
							nil)
					})

					It("returns all warnings", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf([]string{
							"warning-1",
							"warning-2",
							"warning-3",
							"warning-4",
							"warning-5",
							"warning-6",
							"warning-7",
							"warning-8",
							"warning-9",
							"warning-10",
						}))

						Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-org"},
						}}))

						Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-space"},
						}, {
							Filter:   ccv2.OrganizationGUIDFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-org-guid"},
						}}))

						Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(1))
						spaceGUID, queries := fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceArgsForCall(0)
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(queries).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(0))

						Expect(fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupCallCount()).To(Equal(1))
						securityGroupGUID, spaceGUID := fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupArgsForCall(0)
						Expect(securityGroupGUID).To(Equal("some-security-group-guid"))
						Expect(spaceGUID).To(Equal("some-space-guid"))

						Expect(fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupCallCount()).To(Equal(0))
					})
				})
			})

			Context("when the security group is bound to neither running nor staging", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-7", "warning-8"},
						nil,
					)
					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-9", "warning-10"},
						nil,
					)
				})

				Context("when no errors are encountered", func() {
					It("returns all warnings", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf([]string{
							"warning-1",
							"warning-2",
							"warning-3",
							"warning-4",
							"warning-5",
							"warning-6",
							"warning-7",
							"warning-8",
							"warning-9",
							"warning-10",
						}))

						Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(1))
						spaceGUIDStaging, queriesStaging := fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceArgsForCall(0)
						Expect(spaceGUIDStaging).To(Equal("some-space-guid"))
						Expect(queriesStaging).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
						spaceGUIDRunning, queriesRunning := fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)
						Expect(spaceGUIDRunning).To(Equal("some-space-guid"))
						Expect(queriesRunning).To(Equal([]ccv2.QQuery{{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-security-group"},
						}}))

						Expect(fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupCallCount()).To(Equal(0))
						Expect(fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupCallCount()).To(Equal(0))
					})
				})
			})

			Context("when the security group is bound to running", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-7", "warning-8"},
						nil,
					)
					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{
							{
								Name: "some-security-group",
								GUID: "some-security-group-guid",
							},
						},
						ccv2.Warnings{"warning-9", "warning-10"},
						nil,
					)
				})

				It("returns all warnings and a SecurityGroupNotBoundError", func() {
					Expect(warnings).To(ConsistOf([]string{
						"warning-1",
						"warning-2",
						"warning-3",
						"warning-4",
						"warning-5",
						"warning-6",
						"warning-7",
						"warning-8",
						"warning-9",
						"warning-10",
					}))

					Expect(err).To(MatchError(actionerror.SecurityGroupNotBoundError{
						Name:      "some-security-group",
						Lifecycle: lifecycle,
					}))

					Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(1))
					spaceGUIDStaging, queriesStaging := fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceArgsForCall(0)
					Expect(spaceGUIDStaging).To(Equal("some-space-guid"))
					Expect(queriesStaging).To(Equal([]ccv2.QQuery{{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-security-group"},
					}}))

					Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
					spaceGUIDRunning, queriesRunning := fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)
					Expect(spaceGUIDRunning).To(Equal("some-space-guid"))
					Expect(queriesRunning).To(Equal([]ccv2.QQuery{{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-security-group"},
					}}))

					Expect(fakeCloudControllerClient.RemoveSpaceFromStagingSecurityGroupCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.RemoveSpaceFromRunningSecurityGroupCallCount()).To(Equal(0))
				})
			})
		})
	})
})
