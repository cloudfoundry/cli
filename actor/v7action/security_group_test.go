package v7action_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Security Group Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		warnings                  Warnings
		executeErr                error
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil)

	})

	Describe("CreateSecurityGroup", func() {
		const securityGroupName = "security-group-name"
		var (
			filePath              string
			fileContents          []byte
			tempFile              *os.File
			returnedSecurityGroup resources.SecurityGroup
			secGrpPorts           string
			secGrpType            int
			secGrpCode            int
			secGrpDescription     string
			secGrpLog             bool
		)

		BeforeEach(func() {
			fileContents = []byte(`[
	{
		"protocol":"all",
		"destination":"some-destination",
		"ports":"some-ports",
		"type":1,
		"code":0,
		"description":"some-description",
		"log":false
	},
	{
      "protocol": "tcp",
      "destination": "10.10.10.0/24"
    }
]`)
			secGrpPorts = "some-ports"
			secGrpType = 1
			secGrpCode = 0
			secGrpDescription = "some-description"
			secGrpLog = false
			returnedSecurityGroup = resources.SecurityGroup{
				Name: securityGroupName,
				GUID: "some-sec-grp-guid",
				Rules: []resources.Rule{
					{
						Protocol:    "all",
						Destination: "some-destination",
						Ports:       &secGrpPorts,
						Type:        &secGrpType,
						Code:        &secGrpCode,
						Description: &secGrpDescription,
						Log:         &secGrpLog,
					},
					{
						Protocol:    "tcp",
						Destination: "10.10.10.0/24",
					},
				},
			}
			tempFile, executeErr = ioutil.TempFile("", "")
			Expect(executeErr).ToNot(HaveOccurred())
			filePath = tempFile.Name()

			fakeCloudControllerClient.CreateSecurityGroupReturns(returnedSecurityGroup, ccv3.Warnings{"create-sec-grp-warning"}, nil)
		})

		JustBeforeEach(func() {
			_, err := tempFile.Write(fileContents)
			Expect(err).ToNot(HaveOccurred())

			warnings, executeErr = actor.CreateSecurityGroup(securityGroupName, filePath)
		})

		AfterEach(func() {
			os.Remove(filePath)
		})

		When("the path does not exist", func() {
			BeforeEach(func() {
				filePath = "does-not-exist"
			})
			It("returns an error", func() {
				Expect(executeErr).To(HaveOccurred())
				_, ok := executeErr.(*os.PathError)
				Expect(ok).To(BeTrue())
				Expect(warnings).To(Equal(Warnings{}))
			})
		})

		When("Unmarshaling fails", func() {
			BeforeEach(func() {
				fileContents = []byte("not-valid-json")
			})
			It("returns an error", func() {
				Expect(executeErr).To(HaveOccurred())
				_, ok := executeErr.(*json.SyntaxError)
				Expect(ok).To(BeTrue())
				Expect(warnings).To(Equal(Warnings{}))
			})
		})

		It("calls the API with the generated security group resource and returns all warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(warnings).To(Equal(Warnings{"create-sec-grp-warning"}))

			givenSecurityGroup := fakeCloudControllerClient.CreateSecurityGroupArgsForCall(0)

			returnedSecurityGroup.GUID = ""
			Expect(givenSecurityGroup).To(Equal(returnedSecurityGroup))

		})

		When("the security group can't be created", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateSecurityGroupReturns(resources.SecurityGroup{}, ccv3.Warnings{"a-warning"}, errors.New("create-sec-group-error"))
			})
			It("returns the error and warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError("create-sec-group-error"))
				Expect(warnings).To(Equal(Warnings{"a-warning"}))
			})

		})
	})

	Describe("GetSecurityGroup", func() {
		const securityGroupName = "security-group-name"
		var (
			securityGroupSummary SecurityGroupSummary
			description          string
			port                 string
		)

		BeforeEach(func() {
			description = "Top 8 Friends Only"
			port = "9000"
		})

		When("the request succeeds", func() {
			JustBeforeEach(func() {
				securityGroupSummary, warnings, executeErr = actor.GetSecurityGroup(securityGroupName)
				Expect(executeErr).ToNot(HaveOccurred())
			})

			When("the security group is not associated with spaces or rules", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSecurityGroupsReturns(
						[]resources.SecurityGroup{
							{
								GUID:  "some-security-group-guid-1",
								Name:  "some-security-group-1",
								Rules: []resources.Rule{},
							},
						},
						ccv3.Warnings{"warning-1"},
						nil,
					)

					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv3.Space{},
						ccv3.IncludedResources{},
						ccv3.Warnings{"warning-2"},
						nil,
					)
				})

				It("returns the security group summary and warnings", func() {
					Expect(securityGroupSummary).To(Equal(
						SecurityGroupSummary{
							Name:                "security-group-name",
							Rules:               []resources.Rule{},
							SecurityGroupSpaces: []SecurityGroupSpace{},
						},
					))
					Expect(warnings).To(ConsistOf("warning-1"))

					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{"security-group-name"}},
					))

					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(0))
				})
			})

			When("the security group is associated with spaces and rules", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSecurityGroupsReturns(
						[]resources.SecurityGroup{
							{
								GUID: "some-security-group-guid-1",
								Name: "some-security-group-1",
								Rules: []resources.Rule{{
									Destination: "127.0.0.1",
									Description: &description,
									Ports:       &port,
									Protocol:    "tcp",
								}},
								RunningSpaceGUIDs: []string{"space-guid-1"},
								StagingSpaceGUIDs: []string{"space-guid-2"},
							},
						},
						ccv3.Warnings{"warning-1"},
						nil,
					)

					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv3.Space{{
							Name: "my-space",
							GUID: "space-guid-1",
							Relationships: ccv3.Relationships{
								constant.RelationshipTypeOrganization: ccv3.Relationship{GUID: "org-guid-1"},
							},
						}, {
							Name: "your-space",
							GUID: "space-guid-2",
							Relationships: ccv3.Relationships{
								constant.RelationshipTypeOrganization: ccv3.Relationship{GUID: "org-guid-2"},
							},
						}},
						ccv3.IncludedResources{Organizations: []ccv3.Organization{{
							Name: "obsolete-social-networks",
							GUID: "org-guid-1",
						}, {
							Name: "revived-social-networks",
							GUID: "org-guid-2",
						}}},
						ccv3.Warnings{"warning-2"},
						nil,
					)
				})

				It("returns the security group summary and warnings", func() {
					Expect(securityGroupSummary).To(Equal(
						SecurityGroupSummary{
							Name: "security-group-name",
							Rules: []resources.Rule{{
								Destination: "127.0.0.1",
								Description: &description,
								Ports:       &port,
								Protocol:    "tcp",
							}},
							SecurityGroupSpaces: []SecurityGroupSpace{{
								SpaceName: "my-space",
								OrgName:   "obsolete-social-networks",
								Lifecycle: "running",
							}, {
								SpaceName: "your-space",
								OrgName:   "revived-social-networks",
								Lifecycle: "staging",
							}},
						},
					))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{"security-group-name"}},
					))

					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{"space-guid-1", "space-guid-2"}},
						ccv3.Query{Key: ccv3.Include, Values: []string{"organization"}},
					))
				})
			})
		})

		When("the request errors", func() {
			var expectedError error
			JustBeforeEach(func() {
				securityGroupSummary, warnings, executeErr = actor.GetSecurityGroup(securityGroupName)
			})

			When("the security group does not exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSecurityGroupsReturns(
						[]resources.SecurityGroup{},
						ccv3.Warnings{"some-warning"},
						nil,
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ConsistOf("some-warning"))
					Expect(executeErr).To(MatchError(actionerror.SecurityGroupNotFoundError{Name: securityGroupName}))
				})
			})

			When("the cloud controller client errors", func() {
				BeforeEach(func() {
					expectedError = errors.New("I am a CloudControllerClient Error")
					fakeCloudControllerClient.GetSecurityGroupsReturns(
						nil,
						ccv3.Warnings{"some-warning"},
						expectedError,
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ConsistOf("some-warning"))
					Expect(executeErr).To(MatchError(expectedError))
				})
			})
		})
	})

	Describe("GetSecurityGroups", func() {
		var (
			securityGroupSummaries []SecurityGroupSummary
			description            string
			port                   string
		)

		BeforeEach(func() {
			description = "Top 8 Friends Only"
			port = "9000"
		})

		When("the request succeeds", func() {
			JustBeforeEach(func() {
				securityGroupSummaries, warnings, executeErr = actor.GetSecurityGroups()
				Expect(executeErr).ToNot(HaveOccurred())
			})

			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]resources.SecurityGroup{
						{
							GUID: "some-security-group-guid-1",
							Name: "some-security-group-1",
							Rules: []resources.Rule{{
								Destination: "127.0.0.1",
								Description: &description,
								Ports:       &port,
								Protocol:    "tcp",
							}},
							RunningGloballyEnabled: true,
							StagingGloballyEnabled: false,
							RunningSpaceGUIDs:      []string{"space-guid-1"},
							StagingSpaceGUIDs:      []string{"space-guid-2"},
						},
						{
							GUID: "some-security-group-guid-2",
							Name: "some-security-group-2",
							Rules: []resources.Rule{{
								Destination: "127.0.0.1",
								Description: &description,
								Ports:       &port,
								Protocol:    "udp",
							}},
							RunningGloballyEnabled: false,
							StagingGloballyEnabled: true,
							RunningSpaceGUIDs:      []string{"space-guid-2"},
							StagingSpaceGUIDs:      []string{"space-guid-1"},
						},
					},
					ccv3.Warnings{"warning-1"},
					nil,
				)

				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv3.Space{{
						Name: "my-space",
						GUID: "space-guid-1",
						Relationships: ccv3.Relationships{
							constant.RelationshipTypeOrganization: ccv3.Relationship{GUID: "org-guid-1"},
						},
					}, {
						Name: "your-space",
						GUID: "space-guid-2",
						Relationships: ccv3.Relationships{
							constant.RelationshipTypeOrganization: ccv3.Relationship{GUID: "org-guid-2"},
						},
					}},
					ccv3.IncludedResources{Organizations: []ccv3.Organization{{
						Name: "obsolete-social-networks",
						GUID: "org-guid-1",
					}, {
						Name: "revived-social-networks",
						GUID: "org-guid-2",
					}}},
					ccv3.Warnings{"warning-2"},
					nil,
				)
			})

			It("returns the security group summary and warnings", func() {
				Expect(securityGroupSummaries).To(Equal(
					[]SecurityGroupSummary{{
						Name: "some-security-group-1",
						Rules: []resources.Rule{{
							Destination: "127.0.0.1",
							Description: &description,
							Ports:       &port,
							Protocol:    "tcp",
						}},
						SecurityGroupSpaces: []SecurityGroupSpace{{
							SpaceName: "<all>",
							OrgName:   "<all>",
							Lifecycle: "running",
						}, {
							SpaceName: "my-space",
							OrgName:   "obsolete-social-networks",
							Lifecycle: "running",
						}, {
							SpaceName: "your-space",
							OrgName:   "revived-social-networks",
							Lifecycle: "staging",
						}},
					}, {
						Name: "some-security-group-2",
						Rules: []resources.Rule{{
							Destination: "127.0.0.1",
							Description: &description,
							Ports:       &port,
							Protocol:    "udp",
						}},
						SecurityGroupSpaces: []SecurityGroupSpace{{
							SpaceName: "<all>",
							OrgName:   "<all>",
							Lifecycle: "staging",
						}, {
							SpaceName: "your-space",
							OrgName:   "revived-social-networks",
							Lifecycle: "running",
						}, {
							SpaceName: "my-space",
							OrgName:   "obsolete-social-networks",
							Lifecycle: "staging",
						}},
					}},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-2"))

				Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(2))
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{"space-guid-1", "space-guid-2"}},
					ccv3.Query{Key: ccv3.Include, Values: []string{"organization"}},
				))
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(1)).To(ConsistOf(
					ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{"space-guid-2", "space-guid-1"}},
					ccv3.Query{Key: ccv3.Include, Values: []string{"organization"}},
				))
			})
		})

		When("the request errors", func() {
			var expectedError error
			JustBeforeEach(func() {
				securityGroupSummaries, warnings, executeErr = actor.GetSecurityGroups()
			})

			When("there are no security groups", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSecurityGroupsReturns(
						[]resources.SecurityGroup{},
						ccv3.Warnings{"warning-1"},
						nil,
					)
				})

				It("returns an empty list of security group summaries and warnings", func() {
					Expect(securityGroupSummaries).To(Equal(
						[]SecurityGroupSummary{},
					))
					Expect(warnings).To(ConsistOf("warning-1"))

					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(0))
				})
			})

			When("the cloud controller client errors", func() {
				BeforeEach(func() {
					expectedError = errors.New("I am a CloudControllerClient Error")
					fakeCloudControllerClient.GetSecurityGroupsReturns(
						nil,
						ccv3.Warnings{"some-warning"},
						expectedError,
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ConsistOf("some-warning"))
					Expect(executeErr).To(MatchError(expectedError))
				})
			})
		})
	})

	Describe("GetGlobalStagingSecurityGroups", func() {
		var (
			securityGroups []resources.SecurityGroup
		)

		When("the request succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]resources.SecurityGroup{{
						Name: "security-group-name-1",
						GUID: "security-group-guid-1",
					}, {
						Name: "security-group-name-2",
						GUID: "security-group-guid-2",
					}},
					ccv3.Warnings{"security-group-warning"},
					nil,
				)
			})

			JustBeforeEach(func() {
				securityGroups, warnings, executeErr = actor.GetGlobalStagingSecurityGroups()
				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("returns the security groups and warnings", func() {
				Expect(securityGroups).To(Equal(
					[]resources.SecurityGroup{{
						Name: "security-group-name-1",
						GUID: "security-group-guid-1",
					}, {
						Name: "security-group-name-2",
						GUID: "security-group-guid-2",
					}},
				))
				Expect(warnings).To(ConsistOf("security-group-warning"))

				Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.GloballyEnabledStaging, Values: []string{"true"}},
				))
			})
		})

		When("the request errors", func() {
			var expectedError error
			JustBeforeEach(func() {
				securityGroups, warnings, executeErr = actor.GetGlobalStagingSecurityGroups()
			})

			When("there are no security groups", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSecurityGroupsReturns(
						[]resources.SecurityGroup{},
						ccv3.Warnings{"security-group-warning"},
						nil,
					)
				})

				It("returns an empty list of security group summaries and warnings", func() {
					Expect(securityGroups).To(Equal(
						[]resources.SecurityGroup{},
					))
					Expect(warnings).To(ConsistOf("security-group-warning"))

					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.GloballyEnabledStaging, Values: []string{"true"}},
					))
				})
			})

			When("the cloud controller client errors", func() {
				BeforeEach(func() {
					expectedError = errors.New("I am a CloudControllerClient Error")
					fakeCloudControllerClient.GetSecurityGroupsReturns(
						nil,
						ccv3.Warnings{"security-group-warning"},
						expectedError,
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ConsistOf("security-group-warning"))
					Expect(executeErr).To(MatchError(expectedError))
				})
			})
		})
	})
	Describe("GetGlobalRunningingSecurityGroups", func() {
		var (
			securityGroups []resources.SecurityGroup
		)

		When("the request succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]resources.SecurityGroup{{
						Name: "security-group-name-1",
						GUID: "security-group-guid-1",
					}, {
						Name: "security-group-name-2",
						GUID: "security-group-guid-2",
					}},
					ccv3.Warnings{"security-group-warning"},
					nil,
				)
			})

			JustBeforeEach(func() {
				securityGroups, warnings, executeErr = actor.GetGlobalRunningSecurityGroups()
				Expect(executeErr).ToNot(HaveOccurred())
			})

			It("returns the security groups and warnings", func() {
				Expect(securityGroups).To(Equal(
					[]resources.SecurityGroup{{
						Name: "security-group-name-1",
						GUID: "security-group-guid-1",
					}, {
						Name: "security-group-name-2",
						GUID: "security-group-guid-2",
					}},
				))
				Expect(warnings).To(ConsistOf("security-group-warning"))

				Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.GloballyEnabledRunning, Values: []string{"true"}},
				))
			})
		})

		When("the request errors", func() {
			var expectedError error
			JustBeforeEach(func() {
				securityGroups, warnings, executeErr = actor.GetGlobalRunningSecurityGroups()
			})

			When("there are no security groups", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSecurityGroupsReturns(
						[]resources.SecurityGroup{},
						ccv3.Warnings{"security-group-warning"},
						nil,
					)
				})

				It("returns an empty list of security group summaries and warnings", func() {
					Expect(securityGroups).To(Equal(
						[]resources.SecurityGroup{},
					))
					Expect(warnings).To(ConsistOf("security-group-warning"))

					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.GloballyEnabledRunning, Values: []string{"true"}},
					))
				})
			})

			When("the cloud controller client errors", func() {
				BeforeEach(func() {
					expectedError = errors.New("I am a CloudControllerClient Error")
					fakeCloudControllerClient.GetSecurityGroupsReturns(
						nil,
						ccv3.Warnings{"security-group-warning"},
						expectedError,
					)
				})

				It("returns the error and warnings", func() {
					Expect(warnings).To(ConsistOf("security-group-warning"))
					Expect(executeErr).To(MatchError(expectedError))
				})
			})
		})
	})
})
