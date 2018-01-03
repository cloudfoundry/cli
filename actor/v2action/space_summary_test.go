package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Space Summary Actions", func() {
	Describe("GetSpaceSummaryByOrganizationAndName", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
			spaceSummary              SpaceSummary
			warnings                  Warnings
			err                       error
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil)
		})

		Context("when space staging security groups are requested", func() {
			JustBeforeEach(func() {
				spaceSummary, warnings, err = actor.GetSpaceSummaryByOrganizationAndName("some-org-guid", "some-space", true)
			})

			Context("when no errors are encountered", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationReturns(
						ccv2.Organization{
							GUID: "some-org-guid",
							Name: "some-org",
						},
						ccv2.Warnings{"warning-1", "warning-2"},
						nil)

					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{
							{
								GUID: "some-space-guid",
								Name: "some-space",
								SpaceQuotaDefinitionGUID: "some-space-quota-guid",
							},
						},
						ccv2.Warnings{"warning-3", "warning-4"},
						nil)

					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{
							{
								Name: "some-app-2",
							},
							{
								Name: "some-app-1",
							},
						},
						ccv2.Warnings{"warning-5", "warning-6"},
						nil)

					fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
						[]ccv2.ServiceInstance{
							{
								GUID: "some-service-instance-guid-2",
								Name: "some-service-instance-2",
							},
							{
								GUID: "some-service-instance-guid-1",
								Name: "some-service-instance-1",
							},
						},
						ccv2.Warnings{"warning-7", "warning-8"},
						nil)

					fakeCloudControllerClient.GetSpaceQuotaReturns(
						ccv2.SpaceQuota{
							GUID: "some-space-quota-guid",
							Name: "some-space-quota",
						},
						ccv2.Warnings{"warning-9", "warning-10"},
						nil)

					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{
							{
								Name: "some-shared-security-group",
								Rules: []ccv2.SecurityGroupRule{
									{
										Description: "Some shared walking group",
										Destination: "0.0.0.0-5.6.7.8",
										Ports:       "80,443",
										Protocol:    "tcp",
									},
									{
										Description: "Some shared walking group too",
										Destination: "127.10.10.10-127.10.10.255",
										Ports:       "80,4443",
										Protocol:    "udp",
									},
								},
							},
							{
								Name: "some-running-security-group",
								Rules: []ccv2.SecurityGroupRule{
									{
										Description: "Some running walking group",
										Destination: "127.0.0.1-127.0.0.255",
										Ports:       "8080,443",
										Protocol:    "tcp",
									},
									{
										Description: "Some running walking group too",
										Destination: "127.20.20.20-127.20.20.25",
										Ports:       "80,4443",
										Protocol:    "udp",
									},
								},
							},
						},
						ccv2.Warnings{"warning-11", "warning-12"},
						nil)

					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{
							{
								Name: "some-staging-security-group",
								Rules: []ccv2.SecurityGroupRule{
									{
										Description: "Some staging cinematic group",
										Destination: "127.5.5.1-127.6.6.255",
										Ports:       "32767,443",
										Protocol:    "tcp",
									},
									{
										Description: "Some staging cinematic group too",
										Destination: "127.25.20.20-127.25.20.25",
										Ports:       "80,9999",
										Protocol:    "udp",
									},
								},
							},
							{
								Name: "some-shared-security-group",
								Rules: []ccv2.SecurityGroupRule{
									{
										Description: "Some shared cinematic group",
										Destination: "0.0.0.0-5.6.7.8",
										Ports:       "80,443",
										Protocol:    "tcp",
									},
									{
										Description: "Some shared cinematic group too",
										Destination: "127.10.10.10-127.10.10.255",
										Ports:       "80,4443",
										Protocol:    "udp",
									},
								},
							},
						},
						ccv2.Warnings{"warning-13", "warning-14"},
						nil)
				})

				It("returns the space summary and all warnings", func() {
					Expect(err).NotTo(HaveOccurred())

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
						"warning-11",
						"warning-12",
						"warning-13",
						"warning-14",
					}))

					Expect(spaceSummary).To(Equal(SpaceSummary{
						Space: Space{
							Name: "some-space",
							GUID: "some-space-guid",
							SpaceQuotaDefinitionGUID: "some-space-quota-guid",
						},
						OrgName:                   "some-org",
						AppNames:                  []string{"some-app-1", "some-app-2"},
						ServiceInstanceNames:      []string{"some-service-instance-1", "some-service-instance-2"},
						SpaceQuotaName:            "some-space-quota",
						RunningSecurityGroupNames: []string{"some-running-security-group", "some-shared-security-group"},
						StagingSecurityGroupNames: []string{"some-shared-security-group", "some-staging-security-group"},
						SecurityGroupRules: []SecurityGroupRule{
							{
								Name:        "some-running-security-group",
								Description: "Some running walking group",
								Destination: "127.0.0.1-127.0.0.255",
								Lifecycle:   "running",
								Ports:       "8080,443",
								Protocol:    "tcp",
							},
							{
								Name:        "some-running-security-group",
								Description: "Some running walking group too",
								Destination: "127.20.20.20-127.20.20.25",
								Lifecycle:   "running",
								Ports:       "80,4443",
								Protocol:    "udp",
							},
							{
								Name:        "some-shared-security-group",
								Description: "Some shared walking group",
								Destination: "0.0.0.0-5.6.7.8",
								Lifecycle:   "running",
								Ports:       "80,443",
								Protocol:    "tcp",
							},
							{
								Name:        "some-shared-security-group",
								Description: "Some shared cinematic group",
								Destination: "0.0.0.0-5.6.7.8",
								Lifecycle:   "staging",
								Ports:       "80,443",
								Protocol:    "tcp",
							},
							{
								Name:        "some-shared-security-group",
								Description: "Some shared walking group too",
								Destination: "127.10.10.10-127.10.10.255",
								Lifecycle:   "running",
								Ports:       "80,4443",
								Protocol:    "udp",
							},
							{
								Name:        "some-shared-security-group",
								Description: "Some shared cinematic group too",
								Destination: "127.10.10.10-127.10.10.255",
								Lifecycle:   "staging",
								Ports:       "80,4443",
								Protocol:    "udp",
							},
							{
								Name:        "some-staging-security-group",
								Description: "Some staging cinematic group too",
								Destination: "127.25.20.20-127.25.20.25",
								Lifecycle:   "staging",
								Ports:       "80,9999",
								Protocol:    "udp",
							},
							{
								Name:        "some-staging-security-group",
								Description: "Some staging cinematic group",
								Destination: "127.5.5.1-127.6.6.255",
								Lifecycle:   "staging",
								Ports:       "32767,443",
								Protocol:    "tcp",
							},
						},
					}))

					Expect(fakeCloudControllerClient.GetOrganizationCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(0)).To(Equal("some-org-guid"))

					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
					query := fakeCloudControllerClient.GetSpacesArgsForCall(0)
					Expect(query).To(ConsistOf(
						ccv2.QQuery{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-space"},
						},
						ccv2.QQuery{
							Filter:   ccv2.OrganizationGUIDFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-org-guid"},
						},
					))

					Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
					query = fakeCloudControllerClient.GetApplicationsArgsForCall(0)
					Expect(query).To(ConsistOf(
						ccv2.QQuery{
							Filter:   ccv2.SpaceGUIDFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-space-guid"},
						},
					))

					Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))
					spaceGUID, includeUserProvidedServices, query := fakeCloudControllerClient.GetSpaceServiceInstancesArgsForCall(0)
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(includeUserProvidedServices).To(BeTrue())
					Expect(query).To(BeNil())

					Expect(fakeCloudControllerClient.GetSpaceQuotaCallCount()).To(Equal(1))
					spaceQuotaGUID := fakeCloudControllerClient.GetSpaceQuotaArgsForCall(0)
					Expect(spaceQuotaGUID).To(Equal("some-space-quota-guid"))

					Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
					spaceGUIDRunning, queriesRunning := fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)
					Expect(spaceGUIDRunning).To(Equal("some-space-guid"))
					Expect(queriesRunning).To(BeNil())

					Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(1))
					spaceGUIDStaging, queriesStaging := fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceArgsForCall(0)
					Expect(spaceGUIDStaging).To(Equal("some-space-guid"))
					Expect(queriesStaging).To(BeNil())
				})

				Context("when no space quota is assigned", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetSpacesReturns(
							[]ccv2.Space{
								{
									GUID: "some-space-guid",
									Name: "some-space",
								},
							},
							ccv2.Warnings{"warning-3", "warning-4"},
							nil)
					})

					It("does not request space quota information or return a space quota name", func() {
						Expect(fakeCloudControllerClient.GetSpaceQuotaCallCount()).To(Equal(0))
						Expect(spaceSummary.SpaceQuotaName).To(Equal(""))
					})
				})
			})

			Context("when an error is encountered getting the organization", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get-org-error")
					fakeCloudControllerClient.GetOrganizationReturns(
						ccv2.Organization{},
						ccv2.Warnings{
							"warning-1",
							"warning-2",
						},
						expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			Context("when an error is encountered getting the space", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get-space-error")

					fakeCloudControllerClient.GetOrganizationReturns(
						ccv2.Organization{
							GUID: "some-org-guid",
							Name: "some-org",
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{},
						ccv2.Warnings{"warning-1", "warning-2"},
						expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			Context("when an error is encountered getting the application", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get-applications-error")

					fakeCloudControllerClient.GetOrganizationReturns(
						ccv2.Organization{
							GUID: "some-org-guid",
							Name: "some-org",
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{
							{
								GUID: "some-space-guid",
								Name: "some-space",
								SpaceQuotaDefinitionGUID: "some-space-quota-guid",
							},
						},
						nil,
						nil)

					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{},
						ccv2.Warnings{"warning-1", "warning-2"},
						expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			Context("when an error is encountered getting the service instances", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get-service-instances-error")

					fakeCloudControllerClient.GetOrganizationReturns(
						ccv2.Organization{
							GUID: "some-org-guid",
							Name: "some-org",
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{
							{
								GUID: "some-space-guid",
								Name: "some-space",
								SpaceQuotaDefinitionGUID: "some-space-quota-guid",
							},
						},
						nil,
						nil)

					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{
							{
								Name: "some-app-2",
							},
							{
								Name: "some-app-1",
							},
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
						[]ccv2.ServiceInstance{},
						ccv2.Warnings{"warning-1", "warning-2"},
						expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			Context("when an error is encountered getting the space quota", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get-space-quota-error")

					fakeCloudControllerClient.GetOrganizationReturns(
						ccv2.Organization{
							GUID: "some-org-guid",
							Name: "some-org",
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{
							{
								GUID: "some-space-guid",
								Name: "some-space",
								SpaceQuotaDefinitionGUID: "some-space-quota-guid",
							},
						},
						nil,
						nil)

					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{
							{
								Name: "some-app-2",
							},
							{
								Name: "some-app-1",
							},
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
						[]ccv2.ServiceInstance{
							{
								GUID: "some-service-instance-guid-2",
								Name: "some-service-instance-2",
							},
							{
								GUID: "some-service-instance-guid-1",
								Name: "some-service-instance-1",
							},
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpaceQuotaReturns(
						ccv2.SpaceQuota{},
						ccv2.Warnings{"warning-1", "warning-2"},
						expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			Context("when an error is encountered getting the running security groups", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get-running-security-groups-error")

					fakeCloudControllerClient.GetOrganizationReturns(
						ccv2.Organization{
							GUID: "some-org-guid",
							Name: "some-org",
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{
							{
								GUID: "some-space-guid",
								Name: "some-space",
								SpaceQuotaDefinitionGUID: "some-space-quota-guid",
							},
						},
						nil,
						nil)

					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{
							{
								Name: "some-app-2",
							},
							{
								Name: "some-app-1",
							},
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
						[]ccv2.ServiceInstance{
							{
								GUID: "some-service-instance-guid-2",
								Name: "some-service-instance-2",
							},
							{
								GUID: "some-service-instance-guid-1",
								Name: "some-service-instance-1",
							},
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpaceQuotaReturns(
						ccv2.SpaceQuota{
							GUID: "some-space-quota-guid",
							Name: "some-space-quota",
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-1", "warning-2"},
						expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})

			Context("when an error is encountered getting the staging security groups", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get-staging-security-groups-error")

					fakeCloudControllerClient.GetOrganizationReturns(
						ccv2.Organization{
							GUID: "some-org-guid",
							Name: "some-org",
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{
							{
								GUID: "some-space-guid",
								Name: "some-space",
								SpaceQuotaDefinitionGUID: "some-space-quota-guid",
							},
						},
						nil,
						nil)

					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{
							{
								Name: "some-app-2",
							},
							{
								Name: "some-app-1",
							},
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
						[]ccv2.ServiceInstance{
							{
								GUID: "some-service-instance-guid-2",
								Name: "some-service-instance-2",
							},
							{
								GUID: "some-service-instance-guid-1",
								Name: "some-service-instance-1",
							},
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpaceQuotaReturns(
						ccv2.SpaceQuota{
							GUID: "some-space-quota-guid",
							Name: "some-space-quota",
						},
						nil,
						nil)

					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						nil,
						nil)

					fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{},
						ccv2.Warnings{"warning-1", "warning-2"},
						expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})

		Context("when space staging security groups are not requested", func() {
			JustBeforeEach(func() {
				spaceSummary, warnings, err = actor.GetSpaceSummaryByOrganizationAndName("some-org-guid", "some-space", false)
			})

			Context("when no errors are encountered", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationReturns(
						ccv2.Organization{
							GUID: "some-org-guid",
							Name: "some-org",
							DefaultIsolationSegmentGUID: "some-org-default-isolation-segment-guid",
						},
						ccv2.Warnings{"warning-1", "warning-2"},
						nil)

					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv2.Space{
							{
								GUID: "some-space-guid",
								Name: "some-space",
								SpaceQuotaDefinitionGUID: "some-space-quota-guid",
							},
						},
						ccv2.Warnings{"warning-3", "warning-4"},
						nil)

					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{
							{
								Name: "some-app-2",
							},
							{
								Name: "some-app-1",
							},
						},
						ccv2.Warnings{"warning-5", "warning-6"},
						nil)

					fakeCloudControllerClient.GetSpaceServiceInstancesReturns(
						[]ccv2.ServiceInstance{
							{
								GUID: "some-service-instance-guid-2",
								Name: "some-service-instance-2",
							},
							{
								GUID: "some-service-instance-guid-1",
								Name: "some-service-instance-1",
							},
						},
						ccv2.Warnings{"warning-7", "warning-8"},
						nil)

					fakeCloudControllerClient.GetSpaceQuotaReturns(
						ccv2.SpaceQuota{
							GUID: "some-space-quota-guid",
							Name: "some-space-quota",
						},
						ccv2.Warnings{"warning-9", "warning-10"},
						nil)

					fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceReturns(
						[]ccv2.SecurityGroup{
							{
								Name: "some-shared-security-group",
								Rules: []ccv2.SecurityGroupRule{
									{
										Description: "Some shared walking group",
										Destination: "0.0.0.0-5.6.7.8",
										Ports:       "80,443",
										Protocol:    "tcp",
									},
									{
										Description: "Some shared walking group too",
										Destination: "127.10.10.10-127.10.10.255",
										Ports:       "80,4443",
										Protocol:    "udp",
									},
								},
							},
							{
								Name: "some-running-security-group",
								Rules: []ccv2.SecurityGroupRule{
									{
										Description: "Some running walking group",
										Destination: "127.0.0.1-127.0.0.255",
										Ports:       "8080,443",
										Protocol:    "tcp",
									},
									{
										Description: "Some running walking group too",
										Destination: "127.20.20.20-127.20.20.25",
										Ports:       "80,4443",
										Protocol:    "udp",
									},
								},
							},
						},
						ccv2.Warnings{"warning-11", "warning-12"},
						nil)
				})

				It("returns the space summary (without staging security group rules) and all warnings", func() {
					Expect(err).NotTo(HaveOccurred())

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
						"warning-11",
						"warning-12",
					}))

					Expect(spaceSummary).To(Equal(SpaceSummary{
						Space: Space{
							Name: "some-space",
							GUID: "some-space-guid",
							SpaceQuotaDefinitionGUID: "some-space-quota-guid",
						},
						OrgName: "some-org",
						OrgDefaultIsolationSegmentGUID: "some-org-default-isolation-segment-guid",
						AppNames:                       []string{"some-app-1", "some-app-2"},
						ServiceInstanceNames:           []string{"some-service-instance-1", "some-service-instance-2"},
						SpaceQuotaName:                 "some-space-quota",
						RunningSecurityGroupNames:      []string{"some-running-security-group", "some-shared-security-group"},
						StagingSecurityGroupNames:      nil,
						SecurityGroupRules: []SecurityGroupRule{
							{
								Name:        "some-running-security-group",
								Description: "Some running walking group",
								Destination: "127.0.0.1-127.0.0.255",
								Lifecycle:   "running",
								Ports:       "8080,443",
								Protocol:    "tcp",
							},
							{
								Name:        "some-running-security-group",
								Description: "Some running walking group too",
								Destination: "127.20.20.20-127.20.20.25",
								Lifecycle:   "running",
								Ports:       "80,4443",
								Protocol:    "udp",
							},
							{
								Name:        "some-shared-security-group",
								Description: "Some shared walking group",
								Destination: "0.0.0.0-5.6.7.8",
								Lifecycle:   "running",
								Ports:       "80,443",
								Protocol:    "tcp",
							},
							{
								Name:        "some-shared-security-group",
								Description: "Some shared walking group too",
								Destination: "127.10.10.10-127.10.10.255",
								Lifecycle:   "running",
								Ports:       "80,4443",
								Protocol:    "udp",
							},
						},
					}))

					Expect(fakeCloudControllerClient.GetOrganizationCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetOrganizationArgsForCall(0)).To(Equal("some-org-guid"))

					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
					query := fakeCloudControllerClient.GetSpacesArgsForCall(0)
					Expect(query).To(ConsistOf(
						ccv2.QQuery{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-space"},
						},
						ccv2.QQuery{
							Filter:   ccv2.OrganizationGUIDFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-org-guid"},
						},
					))

					Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
					query = fakeCloudControllerClient.GetApplicationsArgsForCall(0)
					Expect(query).To(ConsistOf(
						ccv2.QQuery{
							Filter:   ccv2.SpaceGUIDFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{"some-space-guid"},
						},
					))

					Expect(fakeCloudControllerClient.GetSpaceServiceInstancesCallCount()).To(Equal(1))
					spaceGUID, includeUserProvidedServices, query := fakeCloudControllerClient.GetSpaceServiceInstancesArgsForCall(0)
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(includeUserProvidedServices).To(BeTrue())
					Expect(query).To(BeNil())

					Expect(fakeCloudControllerClient.GetSpaceQuotaCallCount()).To(Equal(1))
					spaceQuotaGUID := fakeCloudControllerClient.GetSpaceQuotaArgsForCall(0)
					Expect(spaceQuotaGUID).To(Equal("some-space-quota-guid"))

					Expect(fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceCallCount()).To(Equal(1))
					spaceGUID, queries := fakeCloudControllerClient.GetSpaceRunningSecurityGroupsBySpaceArgsForCall(0)
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(queries).To(BeNil())

					Expect(fakeCloudControllerClient.GetSpaceStagingSecurityGroupsBySpaceCallCount()).To(Equal(0))
				})
			})
		})
	})
})
