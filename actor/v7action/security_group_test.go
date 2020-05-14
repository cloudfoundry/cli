package v7action_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
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

	Describe("BindSecurityGroupToSpace", func() {
		var (
			lifecycle constant.SecurityGroupLifecycle
			err       error
			warnings  []string
		)

		JustBeforeEach(func() {
			warnings, err = actor.BindSecurityGroupToSpace("some-security-group-guid", "some-space-guid", lifecycle)
		})

		When("the lifecycle is neither running nor staging", func() {
			BeforeEach(func() {
				lifecycle = "bill & ted"
			})

			It("returns and appropriate error", func() {
				Expect(err).To(MatchError(fmt.Sprintf("Invalid lifecycle: %s", lifecycle)))
			})
		})

		When("the lifecycle is running", func() {
			BeforeEach(func() {
				lifecycle = constant.SecurityGroupLifecycleRunning
			})

			When("binding the space does not return an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateSecurityGroupRunningSpaceReturns(
						ccv3.Warnings{"warning-1"},
						nil,
					)
				})

				It("returns warnings and no error", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1"))
					Expect(fakeCloudControllerClient.UpdateSecurityGroupRunningSpaceCallCount()).To(Equal(1))
					securityGroupGUID, spaceGUID := fakeCloudControllerClient.UpdateSecurityGroupRunningSpaceArgsForCall(0)
					Expect(securityGroupGUID).To(Equal("some-security-group-guid"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			When("binding the space returns an error", func() {
				var returnedError error
				BeforeEach(func() {
					returnedError = errors.New("associate-space-error")
					fakeCloudControllerClient.UpdateSecurityGroupRunningSpaceReturns(
						ccv3.Warnings{"warning-1", "warning-2"},
						returnedError,
					)
				})

				It("returns the error and warnings", func() {
					Expect(err).To(Equal(returnedError))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})

		When("the lifecycle is staging", func() {
			BeforeEach(func() {
				lifecycle = constant.SecurityGroupLifecycleStaging
			})

			When("binding the space does not return an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateSecurityGroupStagingSpaceReturns(
						ccv3.Warnings{"warning-1", "warning-2"},
						nil,
					)
				})

				It("returns warnings and no error", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(fakeCloudControllerClient.UpdateSecurityGroupStagingSpaceCallCount()).To(Equal(1))
					securityGroupGUID, spaceGUID := fakeCloudControllerClient.UpdateSecurityGroupStagingSpaceArgsForCall(0)
					Expect(securityGroupGUID).To(Equal("some-security-group-guid"))
					Expect(spaceGUID).To(Equal("some-space-guid"))
				})
			})

			When("binding the space returns an error", func() {
				var returnedError error
				BeforeEach(func() {
					returnedError = errors.New("associate-space-error")
					fakeCloudControllerClient.UpdateSecurityGroupStagingSpaceReturns(
						ccv3.Warnings{"warning-1", "warning-2"},
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

		When("the json does not contain the required fields", func() {
			BeforeEach(func() {
				fileContents = []byte(`{"blah": "blah"}`)
			})
			It("returns an error", func() {
				Expect(executeErr).To(HaveOccurred())
				_, ok := executeErr.(*json.UnmarshalTypeError)
				Expect(ok).To(BeTrue())
				Expect(warnings).To(Equal(Warnings{}))
			})
		})
		When("the json is invalid", func() {
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

	Describe("GetSecurityGroupSummary", func() {
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
				securityGroupSummary, warnings, executeErr = actor.GetSecurityGroupSummary(securityGroupName)
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
							Relationships: resources.Relationships{
								constant.RelationshipTypeOrganization: resources.Relationship{GUID: "org-guid-1"},
							},
						}, {
							Name: "your-space",
							GUID: "space-guid-2",
							Relationships: resources.Relationships{
								constant.RelationshipTypeOrganization: resources.Relationship{GUID: "org-guid-2"},
							},
						}},
						ccv3.IncludedResources{Organizations: []resources.Organization{{
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
				securityGroupSummary, warnings, executeErr = actor.GetSecurityGroupSummary(securityGroupName)
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
			trueVal                bool
			falseVal               bool
		)

		BeforeEach(func() {
			description = "Top 8 Friends Only"
			port = "9000"
			trueVal = true
			falseVal = false
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
							RunningGloballyEnabled: &trueVal,
							StagingGloballyEnabled: &falseVal,
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
							RunningGloballyEnabled: &falseVal,
							StagingGloballyEnabled: &trueVal,
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
						Relationships: resources.Relationships{
							constant.RelationshipTypeOrganization: resources.Relationship{GUID: "org-guid-1"},
						},
					}, {
						Name: "your-space",
						GUID: "space-guid-2",
						Relationships: resources.Relationships{
							constant.RelationshipTypeOrganization: resources.Relationship{GUID: "org-guid-2"},
						},
					}},
					ccv3.IncludedResources{Organizations: []resources.Organization{{
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

	Describe("GetSecurityGroup", func() {
		var (
			securityGroupName = "tom"
			securityGroup     resources.SecurityGroup
		)

		When("the request succeeds", func() {
			JustBeforeEach(func() {
				securityGroup, warnings, executeErr = actor.GetSecurityGroup(securityGroupName)
				Expect(executeErr).ToNot(HaveOccurred())
			})

			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]resources.SecurityGroup{
						{
							GUID: "some-security-group-guid-1",
							Name: "some-security-group-1",
						},
					},
					ccv3.Warnings{"warning-1"},
					nil,
				)
			})

			It("returns the security group and warnings", func() {
				Expect(securityGroup).To(Equal(
					resources.SecurityGroup{
						Name: "some-security-group-1",
						GUID: "some-security-group-guid-1",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1"))

				Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
			})
		})

		When("the request errors", func() {
			var expectedError error

			JustBeforeEach(func() {
				securityGroup, warnings, executeErr = actor.GetSecurityGroup(securityGroupName)
			})

			When("there are no security groups", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSecurityGroupsReturns(
						[]resources.SecurityGroup{},
						ccv3.Warnings{"warning-1"},
						nil,
					)
				})

				It("returns an empty security group and warnings", func() {
					Expect(securityGroup).To(Equal(resources.SecurityGroup{}))
					Expect(warnings).To(ConsistOf("warning-1"))

					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
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

	Describe("UnbindSecurityGroup", func() {
		var (
			securityGroupName = "some-security-group"
			orgName           = "some-org"
			spaceName         = "some-space"
			lifecycle         = constant.SecurityGroupLifecycleStaging
			warnings          Warnings
			executeErr        error
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.UnbindSecurityGroup(securityGroupName, orgName, spaceName, lifecycle)
		})

		BeforeEach(func() {
			fakeCloudControllerClient.GetOrganizationsReturns(
				[]resources.Organization{{GUID: "some-org-guid"}},
				ccv3.Warnings{"get-org-warning"},
				nil,
			)

			fakeCloudControllerClient.GetSpacesReturns(
				[]ccv3.Space{{GUID: "some-space-guid"}},
				ccv3.IncludedResources{},
				ccv3.Warnings{"get-space-warning"},
				nil,
			)

			fakeCloudControllerClient.GetSecurityGroupsReturns(
				[]resources.SecurityGroup{{GUID: "some-security-group-guid"}},
				ccv3.Warnings{"get-security-group-warning"},
				nil,
			)

			fakeCloudControllerClient.UnbindSecurityGroupStagingSpaceReturns(
				ccv3.Warnings{"unbind-security-group-warning"},
				nil,
			)
		})

		When("all requests succeed", func() {
			It("returns warnings and no error", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-org-warning", "get-space-warning", "get-security-group-warning", "unbind-security-group-warning"))

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				orgsQuery := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(orgsQuery).To(Equal([]ccv3.Query{
					{Key: ccv3.NameFilter, Values: []string{orgName}},
				}))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				spacesQuery := fakeCloudControllerClient.GetSpacesArgsForCall(0)
				Expect(spacesQuery).To(Equal([]ccv3.Query{
					{Key: ccv3.NameFilter, Values: []string{spaceName}},
					{Key: ccv3.OrganizationGUIDFilter, Values: []string{"some-org-guid"}},
				}))

				Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
				securityGroupsQuery := fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)
				Expect(securityGroupsQuery).To(Equal([]ccv3.Query{
					{Key: ccv3.NameFilter, Values: []string{securityGroupName}},
				}))

				Expect(fakeCloudControllerClient.UnbindSecurityGroupStagingSpaceCallCount()).To(Equal(1))
				givenSecurityGroupGUID, givenSpaceGUID := fakeCloudControllerClient.UnbindSecurityGroupStagingSpaceArgsForCall(0)
				Expect(givenSecurityGroupGUID).To(Equal("some-security-group-guid"))
				Expect(givenSpaceGUID).To(Equal("some-space-guid"))
			})
		})

		When("the seurity group is not bound to the space", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UnbindSecurityGroupStagingSpaceReturns(
					ccv3.Warnings{"get-security-group-warning"},
					ccerror.SecurityGroupNotBound{},
				)
			})

			It("returns warnings and an appropriate error", func() {
				Expect(executeErr).To(MatchError(actionerror.SecurityGroupNotBoundToSpaceError{
					Name:      securityGroupName,
					Space:     spaceName,
					Lifecycle: lifecycle,
				}))
			})
		})

		When("getting the org fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]resources.Organization{{}},
					ccv3.Warnings{"get-org-warning"},
					errors.New("org error"),
				)
			})
			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("org error"))
			})
		})

		When("getting the space fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv3.Space{{}},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-space-warning"},
					errors.New("space error"),
				)
			})
			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("space error"))
			})
		})

		When("getting the security group fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]resources.SecurityGroup{{}},
					ccv3.Warnings{"get-security-group-warning"},
					errors.New("security group error"),
				)
			})
			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("security group error"))
			})
		})

		When("binding the security group fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UnbindSecurityGroupStagingSpaceReturns(
					ccv3.Warnings{"get-security-group-warning"},
					errors.New("security group unbind error"),
				)
			})
			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("security group unbind error"))
			})
		})
	})

	Describe("UpdateSecurityGroup", func() {
		const securityGroupName = "security-group-name"
		var (
			filePath              string
			fileContents          []byte
			tempFile              *os.File
			originalSecurityGroup resources.SecurityGroup
			updatedSecurityGroup  resources.SecurityGroup
		)

		BeforeEach(func() {
			originalSecurityGroup = resources.SecurityGroup{
				Name:  securityGroupName,
				GUID:  "some-sec-grp-guid",
				Rules: []resources.Rule{},
			}
			fakeCloudControllerClient.GetSecurityGroupsReturns(
				[]resources.SecurityGroup{originalSecurityGroup},
				ccv3.Warnings{"get-security-group-warning"},
				nil,
			)

			fileContents = []byte(`[
	{
      "protocol": "tcp",
      "destination": "10.10.10.0/24"
    }
]`)
			tempFile, executeErr = ioutil.TempFile("", "")
			Expect(executeErr).ToNot(HaveOccurred())
			filePath = tempFile.Name()

			updatedSecurityGroup = resources.SecurityGroup{
				Name: securityGroupName,
				GUID: "some-sec-grp-guid",
				Rules: []resources.Rule{
					{
						Protocol:    "tcp",
						Destination: "10.10.10.0/24",
					},
				},
			}
			fakeCloudControllerClient.UpdateSecurityGroupReturns(updatedSecurityGroup, ccv3.Warnings{"update-warning"}, nil)
		})

		JustBeforeEach(func() {
			_, err := tempFile.Write(fileContents)
			Expect(err).ToNot(HaveOccurred())

			warnings, executeErr = actor.UpdateSecurityGroup(securityGroupName, filePath)
		})

		AfterEach(func() {
			os.Remove(filePath)
		})

		It("parses the input file, finds the requested security group, and updates it", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(warnings).To(Equal(Warnings{"get-security-group-warning", "update-warning"}))

			givenQuery := fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)
			Expect(givenQuery).To(ConsistOf(
				ccv3.Query{Key: ccv3.NameFilter, Values: []string{securityGroupName}},
			))

			givenSecurityGroup := fakeCloudControllerClient.UpdateSecurityGroupArgsForCall(0)
			Expect(givenSecurityGroup).To(Equal(updatedSecurityGroup))
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

		When("the security group does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]resources.SecurityGroup{},
					ccv3.Warnings{"get-security-group-warning"},
					nil,
				)
			})

			It("returns a security group not found error and warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError(actionerror.SecurityGroupNotFoundError{Name: securityGroupName}))
				Expect(warnings).To(Equal(Warnings{"get-security-group-warning"}))
			})
		})

		When("the security group can't be updated", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateSecurityGroupReturns(
					resources.SecurityGroup{},
					ccv3.Warnings{"a-warning"},
					errors.New("update-error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError("update-error"))
				Expect(warnings).To(Equal(Warnings{"get-security-group-warning", "a-warning"}))
			})
		})
	})

	Describe("UpdateSecurityGroupGloballyEnabled", func() {
		var (
			securityGroupName = "tom"
			globallyEnabled   bool
			lifeycle          constant.SecurityGroupLifecycle
			executeErr        error

			trueValue  = true
			falseValue = false
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateSecurityGroupGloballyEnabled(securityGroupName, lifeycle, globallyEnabled)
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]resources.SecurityGroup{
						{
							GUID: "some-security-group-guid",
							Name: securityGroupName,
						},
					},
					ccv3.Warnings{"warning-1"},
					nil,
				)

				fakeCloudControllerClient.UpdateSecurityGroupReturns(
					resources.SecurityGroup{},
					ccv3.Warnings{"warning-2"},
					nil,
				)
			})

			When("updating staging to true", func() {
				BeforeEach(func() {
					lifeycle = constant.SecurityGroupLifecycleStaging
					globallyEnabled = true
				})

				It("returns the warnings", func() {
					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					query := fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)
					Expect(query).To(Equal([]ccv3.Query{
						{Key: ccv3.NameFilter, Values: []string{securityGroupName}},
					}))

					Expect(fakeCloudControllerClient.UpdateSecurityGroupCallCount()).To(Equal(1))
					args := fakeCloudControllerClient.UpdateSecurityGroupArgsForCall(0)
					Expect(args).To(Equal(resources.SecurityGroup{
						GUID:                   "some-security-group-guid",
						StagingGloballyEnabled: &trueValue,
					}))

					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(executeErr).NotTo(HaveOccurred())
				})
			})

			When("updating staging to false", func() {
				BeforeEach(func() {
					lifeycle = constant.SecurityGroupLifecycleStaging
					globallyEnabled = false
				})

				It("returns the warnings", func() {
					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					query := fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)
					Expect(query).To(Equal([]ccv3.Query{
						{Key: ccv3.NameFilter, Values: []string{securityGroupName}},
					}))

					Expect(fakeCloudControllerClient.UpdateSecurityGroupCallCount()).To(Equal(1))
					args := fakeCloudControllerClient.UpdateSecurityGroupArgsForCall(0)
					Expect(args).To(Equal(resources.SecurityGroup{
						GUID:                   "some-security-group-guid",
						StagingGloballyEnabled: &falseValue,
					}))

					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(executeErr).NotTo(HaveOccurred())
				})
			})

			When("updating running to true", func() {
				BeforeEach(func() {
					lifeycle = constant.SecurityGroupLifecycleRunning
					globallyEnabled = true
				})

				It("returns the warnings", func() {
					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					query := fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)
					Expect(query).To(Equal([]ccv3.Query{
						{Key: ccv3.NameFilter, Values: []string{securityGroupName}},
					}))

					Expect(fakeCloudControllerClient.UpdateSecurityGroupCallCount()).To(Equal(1))
					args := fakeCloudControllerClient.UpdateSecurityGroupArgsForCall(0)
					Expect(args).To(Equal(resources.SecurityGroup{
						GUID:                   "some-security-group-guid",
						RunningGloballyEnabled: &trueValue,
					}))

					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(executeErr).NotTo(HaveOccurred())
				})
			})

			When("updating running to false", func() {
				BeforeEach(func() {
					lifeycle = constant.SecurityGroupLifecycleRunning
					globallyEnabled = false
				})

				It("returns the warnings", func() {
					Expect(fakeCloudControllerClient.GetSecurityGroupsCallCount()).To(Equal(1))
					query := fakeCloudControllerClient.GetSecurityGroupsArgsForCall(0)
					Expect(query).To(Equal([]ccv3.Query{
						{Key: ccv3.NameFilter, Values: []string{securityGroupName}},
					}))

					Expect(fakeCloudControllerClient.UpdateSecurityGroupCallCount()).To(Equal(1))
					args := fakeCloudControllerClient.UpdateSecurityGroupArgsForCall(0)
					Expect(args).To(Equal(resources.SecurityGroup{
						GUID:                   "some-security-group-guid",
						RunningGloballyEnabled: &falseValue,
					}))

					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(executeErr).NotTo(HaveOccurred())
				})
			})
		})

		When("the request to get the security group errors", func() {
			BeforeEach(func() {
				lifeycle = constant.SecurityGroupLifecycleRunning
				globallyEnabled = false

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					nil,
					ccv3.Warnings{"warning-1"},
					errors.New("get-group-error"),
				)

				fakeCloudControllerClient.UpdateSecurityGroupReturns(
					resources.SecurityGroup{},
					ccv3.Warnings{"warning-2"},
					nil,
				)
			})

			It("returns the warnings and error", func() {
				Expect(warnings).To(ConsistOf("warning-1"))
				Expect(executeErr).To(MatchError("get-group-error"))
			})
		})

		When("the request to update the security group errors", func() {
			BeforeEach(func() {
				lifeycle = constant.SecurityGroupLifecycleRunning
				globallyEnabled = false

				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]resources.SecurityGroup{
						{
							GUID: "some-security-group-guid",
							Name: securityGroupName,
						},
					},
					ccv3.Warnings{"warning-1"},
					nil,
				)

				fakeCloudControllerClient.UpdateSecurityGroupReturns(
					resources.SecurityGroup{},
					ccv3.Warnings{"warning-2"},
					errors.New("update-group-error"),
				)
			})

			It("returns the warnings and error", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(executeErr).To(MatchError("update-group-error"))
			})
		})
	})

	Describe("DeleteSecurityGroup", func() {
		var securityGroupName = "elsa"

		JustBeforeEach(func() {
			warnings, executeErr = actor.DeleteSecurityGroup(securityGroupName)
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]resources.SecurityGroup{
						{
							GUID: "some-security-group-guid",
							Name: securityGroupName,
						},
					},
					ccv3.Warnings{"warning-1"},
					nil,
				)

				fakeCloudControllerClient.DeleteSecurityGroupReturns(
					ccv3.JobURL("https://jobs/job_guid"),
					ccv3.Warnings{"warning-2"},
					nil,
				)

				fakeCloudControllerClient.PollJobReturns(
					ccv3.Warnings{"warning-3"},
					nil,
				)
			})

			It("deletes the security group asynchronously", func() {
				// Delete the security group asynchronously
				Expect(fakeCloudControllerClient.DeleteSecurityGroupCallCount()).To(Equal(1))
				passedSecurityGroupGuid := fakeCloudControllerClient.DeleteSecurityGroupArgsForCall(0)
				Expect(passedSecurityGroupGuid).To(Equal("some-security-group-guid"))

				// Poll the delete job
				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
				responseJobUrl := fakeCloudControllerClient.PollJobArgsForCall(0)
				Expect(responseJobUrl).To(Equal(ccv3.JobURL("https://jobs/job_guid")))
			})

			It("returns the warnings and error", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3"))
				Expect(executeErr).To(BeNil())
			})
		})

		When("the request to get the security group errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					nil,
					ccv3.Warnings{"warning-1"},
					errors.New("get-group-error"),
				)

				fakeCloudControllerClient.DeleteSecurityGroupReturns(
					ccv3.JobURL(""),
					ccv3.Warnings{"warning-2"},
					nil,
				)
			})

			It("returns the warnings and error", func() {
				Expect(warnings).To(ConsistOf("warning-1"))
				Expect(executeErr).To(MatchError("get-group-error"))
			})
		})

		When("the request to delete the security group errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSecurityGroupsReturns(
					[]resources.SecurityGroup{
						{
							GUID: "some-security-group-guid",
							Name: securityGroupName,
						},
					},
					ccv3.Warnings{"warning-1"},
					nil,
				)

				fakeCloudControllerClient.DeleteSecurityGroupReturns(
					ccv3.JobURL(""),
					ccv3.Warnings{"warning-2"},
					errors.New("delete-group-error"),
				)
			})

			It("returns the warnings and error", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(executeErr).To(MatchError("delete-group-error"))
			})
		})
	})
})
