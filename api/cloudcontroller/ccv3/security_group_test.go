package ccv3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/resources"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SecurityGroup", func() {
	var (
		client    *Client
		requester *ccv3fakes.FakeRequester
	)

	BeforeEach(func() {
		requester = new(ccv3fakes.FakeRequester)
		client, _ = NewFakeRequesterTestClient(requester)
	})

	Describe("CreateSecurityGroup", func() {
		var (
			securityGroupName    string
			securityGroupParams  resources.SecurityGroup
			createdSecurityGroup resources.SecurityGroup
			warnings             Warnings
			executeErr           error
		)

		BeforeEach(func() {
			securityGroupName = "some-group-name"
			requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
				requestParams.ResponseBody.(*resources.SecurityGroup).GUID = "some-guid"
				return "", Warnings{"some-warning"}, errors.New("some-error")
			})
			securityGroupParams = resources.SecurityGroup{
				Name: securityGroupName,
				Rules: []resources.Rule{
					{
						Protocol:    "tcp",
						Destination: "10.0.10.0/24",
					},
				},
			}
		})

		JustBeforeEach(func() {
			createdSecurityGroup, warnings, executeErr = client.CreateSecurityGroup(securityGroupParams)
		})

		It("makes the correct request", func() {
			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			actualParams := requester.MakeRequestArgsForCall(0)
			Expect(actualParams.RequestName).To(Equal(internal.PostSecurityGroupRequest))
			Expect(actualParams.RequestBody).To(Equal(securityGroupParams))
			Expect(actualParams.ResponseBody).To(HaveTypeOf(&resources.SecurityGroup{}))
		})

		It("returns the given role and all warnings", func() {
			Expect(createdSecurityGroup).To(Equal(resources.SecurityGroup{GUID: "some-guid"}))
			Expect(warnings).To(ConsistOf("some-warning"))
			Expect(executeErr).To(MatchError("some-error"))
		})
	})

	Describe("GetSecurityGroups", func() {
		var (
			returnedSecurityGroups []resources.SecurityGroup
			query                  = Query{}
			warnings               Warnings
			executeErr             error
		)

		BeforeEach(func() {
			requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
				err := requestParams.AppendToList(resources.SecurityGroup{Name: "security-group-name-1", GUID: "security-group-guid-1"})
				Expect(err).NotTo(HaveOccurred())
				return IncludedResources{}, Warnings{"some-warning"}, errors.New("some-error")
			})
		})

		JustBeforeEach(func() {
			returnedSecurityGroups, warnings, executeErr = client.GetSecurityGroups(query)
		})

		It("makes the correct request", func() {
			Expect(requester.MakeListRequestCallCount()).To(Equal(1))
			params := requester.MakeListRequestArgsForCall(0)

			Expect(params.RequestName).To(Equal(internal.GetSecurityGroupsRequest))
			Expect(params.Query).To(Equal([]Query{query}))
			Expect(params.ResponseBody).To(Equal(resources.SecurityGroup{}))
		})

		It("returns the resources and all warnings", func() {
			Expect(warnings).To(ConsistOf("some-warning"))
			Expect(executeErr).To(MatchError("some-error"))
			Expect(returnedSecurityGroups).To(Equal([]resources.SecurityGroup{{
				GUID: "security-group-guid-1",
				Name: "security-group-name-1",
			}}))
		})
	})

	Describe("GetRunningSecurityGroups", func() {
		var (
			spaceGUID              = "some-space-guid"
			returnedSecurityGroups []resources.SecurityGroup
			query                  Query
			warnings               Warnings
			executeErr             error
		)

		BeforeEach(func() {
			requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
				err := requestParams.AppendToList(resources.SecurityGroup{Name: "security-group-name-1", GUID: "security-group-guid-1"})
				Expect(err).NotTo(HaveOccurred())
				return IncludedResources{}, Warnings{"some-warning"}, errors.New("some-error")
			})
		})

		JustBeforeEach(func() {
			returnedSecurityGroups, warnings, executeErr = client.GetRunningSecurityGroups(spaceGUID, query)
		})

		It("makes the correct request", func() {
			Expect(requester.MakeListRequestCallCount()).To(Equal(1))
			params := requester.MakeListRequestArgsForCall(0)

			Expect(params.RequestName).To(Equal(internal.GetSpaceRunningSecurityGroupsRequest))
			Expect(params.URIParams).To(Equal(internal.Params{"space_guid": spaceGUID}))
			Expect(params.Query).To(Equal([]Query{query}))
			Expect(params.ResponseBody).To(Equal(resources.SecurityGroup{}))
		})

		It("returns the resources and all warnings", func() {
			Expect(executeErr).To(MatchError("some-error"))
			Expect(warnings).To(Equal(Warnings{"some-warning"}))
			Expect(returnedSecurityGroups).To(Equal([]resources.SecurityGroup{{
				GUID: "security-group-guid-1",
				Name: "security-group-name-1",
			}}))
		})
	})

	Describe("GetStagingSecurityGroups", func() {
		var (
			spaceGUID              = "some-space-guid"
			returnedSecurityGroups []resources.SecurityGroup
			query                  Query
			warnings               Warnings
			executeErr             error
		)

		BeforeEach(func() {
			requester.MakeListRequestCalls(func(requestParams RequestParams) (IncludedResources, Warnings, error) {
				err := requestParams.AppendToList(resources.SecurityGroup{Name: "security-group-name-1", GUID: "security-group-guid-1"})
				Expect(err).NotTo(HaveOccurred())
				return IncludedResources{}, Warnings{"some-warning"}, errors.New("some-error")
			})
		})

		JustBeforeEach(func() {
			returnedSecurityGroups, warnings, executeErr = client.GetStagingSecurityGroups(spaceGUID, query)
		})

		It("makes the correct request", func() {
			Expect(requester.MakeListRequestCallCount()).To(Equal(1))
			params := requester.MakeListRequestArgsForCall(0)

			Expect(params.RequestName).To(Equal(internal.GetSpaceStagingSecurityGroupsRequest))
			Expect(params.URIParams).To(Equal(internal.Params{"space_guid": spaceGUID}))
			Expect(params.Query).To(Equal([]Query{query}))
			Expect(params.ResponseBody).To(Equal(resources.SecurityGroup{}))
		})

		It("returns the resources and all warnings", func() {
			Expect(executeErr).To(MatchError("some-error"))
			Expect(warnings).To(Equal(Warnings{"some-warning"}))
			Expect(returnedSecurityGroups).To(Equal([]resources.SecurityGroup{{
				GUID: "security-group-guid-1",
				Name: "security-group-name-1",
			}}))
		})
	})

	Describe("UnbindSecurityGroupRunningSpace", func() {
		var (
			spaceGUID         = "some-space-guid"
			securityGroupGUID = "some-security-group-guid"
			warnings          Warnings
			executeErr        error
		)

		BeforeEach(func() {
			requester.MakeRequestReturns(JobURL(""), Warnings{"some-warning"}, errors.New("some-error"))
		})

		JustBeforeEach(func() {
			warnings, executeErr = client.UnbindSecurityGroupRunningSpace(securityGroupGUID, spaceGUID)
		})

		It("makes the correct request", func() {
			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			params := requester.MakeRequestArgsForCall(0)

			Expect(params.RequestName).To(Equal(internal.DeleteSecurityGroupRunningSpaceRequest))
			Expect(params.URIParams).To(Equal(internal.Params{
				"security_group_guid": securityGroupGUID, "space_guid": spaceGUID,
			}))
		})

		It("returns the resources and all warnings", func() {
			Expect(executeErr).To(MatchError("some-error"))
			Expect(warnings).To(Equal(Warnings{"some-warning"}))
		})
	})

	Describe("UnbindSecurityGroupStagingSpace", func() {
		var (
			spaceGUID         = "some-space-guid"
			securityGroupGUID = "some-security-group-guid"
			warnings          Warnings
			executeErr        error
		)

		BeforeEach(func() {
			requester.MakeRequestReturns(JobURL(""), Warnings{"some-warning"}, errors.New("some-error"))
		})

		JustBeforeEach(func() {
			warnings, executeErr = client.UnbindSecurityGroupStagingSpace(securityGroupGUID, spaceGUID)
		})

		It("makes the correct request", func() {
			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			params := requester.MakeRequestArgsForCall(0)

			Expect(params.RequestName).To(Equal(internal.DeleteSecurityGroupStagingSpaceRequest))
			Expect(params.URIParams).To(Equal(internal.Params{
				"security_group_guid": securityGroupGUID, "space_guid": spaceGUID,
			}))
		})

		It("returns the resources and all warnings", func() {
			Expect(executeErr).To(MatchError("some-error"))
			Expect(warnings).To(Equal(Warnings{"some-warning"}))
		})
	})

	Describe("UpdateSecurityGroupRunningSpace", func() {
		var (
			spaceGUID         = "some-space-guid"
			securityGroupGUID = "some-security-group-guid"
			warnings          Warnings
			executeErr        error
		)

		BeforeEach(func() {
			requester.MakeRequestReturns(JobURL(""), Warnings{"some-warning"}, errors.New("some-error"))
		})

		JustBeforeEach(func() {
			warnings, executeErr = client.UpdateSecurityGroupRunningSpace(securityGroupGUID, spaceGUID)
		})

		It("makes the correct request", func() {
			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			params := requester.MakeRequestArgsForCall(0)

			Expect(params.RequestName).To(Equal(internal.PostSecurityGroupRunningSpaceRequest))
			Expect(params.URIParams).To(Equal(internal.Params{"security_group_guid": securityGroupGUID}))
			Expect(params.RequestBody).To(Equal(RelationshipList{
				GUIDs: []string{spaceGUID},
			}))
		})

		It("returns the resources and all warnings", func() {
			Expect(executeErr).To(MatchError("some-error"))
			Expect(warnings).To(Equal(Warnings{"some-warning"}))
		})
	})

	Describe("UpdateSecurityGroupStagingSpace", func() {
		var (
			spaceGUID         = "some-space-guid"
			securityGroupGUID = "some-security-group-guid"
			warnings          Warnings
			executeErr        error
		)

		BeforeEach(func() {
			requester.MakeRequestReturns(JobURL(""), Warnings{"some-warning"}, errors.New("some-error"))
		})

		JustBeforeEach(func() {
			warnings, executeErr = client.UpdateSecurityGroupStagingSpace(securityGroupGUID, spaceGUID)
		})

		It("makes the correct request", func() {
			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			params := requester.MakeRequestArgsForCall(0)

			Expect(params.RequestName).To(Equal(internal.PostSecurityGroupStagingSpaceRequest))
			Expect(params.URIParams).To(Equal(internal.Params{"security_group_guid": securityGroupGUID}))
			Expect(params.RequestBody).To(Equal(RelationshipList{
				GUIDs: []string{spaceGUID},
			}))
		})

		It("returns the resources and all warnings", func() {
			Expect(executeErr).To(MatchError("some-error"))
			Expect(warnings).To(Equal(Warnings{"some-warning"}))
		})
	})

	Describe("UpdateSecurityGroup", func() {
		var (
			securityGroup         resources.SecurityGroup
			returnedSecurityGroup resources.SecurityGroup
			warnings              Warnings
			executeErr            error
			trueValue             = true
		)

		BeforeEach(func() {
			requester.MakeRequestCalls(func(requestParams RequestParams) (JobURL, Warnings, error) {
				requestParams.ResponseBody.(*resources.SecurityGroup).GUID = "returned-group-guid"
				return "", Warnings{"some-warning"}, errors.New("some-error")
			})

			securityGroup = resources.SecurityGroup{
				Name:                   "some-security-group-name",
				GUID:                   "some-security-group-guid",
				StagingGloballyEnabled: &trueValue,
			}
		})

		JustBeforeEach(func() {
			returnedSecurityGroup, warnings, executeErr = client.UpdateSecurityGroup(securityGroup)
		})

		It("makes the correct request", func() {
			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			params := requester.MakeRequestArgsForCall(0)

			Expect(params.RequestName).To(Equal(internal.PatchSecurityGroupRequest))
			Expect(params.URIParams).To(Equal(internal.Params{"security_group_guid": securityGroup.GUID}))
			Expect(params.RequestBody).To(Equal(resources.SecurityGroup{
				StagingGloballyEnabled: &trueValue,
			}))
		})

		It("returns the resource and all warnings", func() {
			Expect(returnedSecurityGroup).To(Equal(resources.SecurityGroup{
				GUID: "returned-group-guid",
			}))
			Expect(executeErr).To(MatchError("some-error"))
			Expect(warnings).To(Equal(Warnings{"some-warning"}))
		})
	})

	Describe("DeleteSecurityGroup", func() {
		var (
			securityGroupGUID = "some-security-group-guid"
			jobURL            JobURL
			warnings          Warnings
			executeErr        error
		)

		BeforeEach(func() {
			requester.MakeRequestReturns(JobURL("some-job-url"), Warnings{"some-warning"}, errors.New("some-error"))
		})

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.DeleteSecurityGroup(securityGroupGUID)
		})

		It("makes the correct request", func() {
			Expect(requester.MakeRequestCallCount()).To(Equal(1))
			params := requester.MakeRequestArgsForCall(0)

			Expect(params.RequestName).To(Equal(internal.DeleteSecurityGroupRequest))
			Expect(params.URIParams).To(Equal(internal.Params{"security_group_guid": securityGroupGUID}))
		})

		It("returns all warnings and a job url", func() {
			Expect(executeErr).To(MatchError("some-error"))
			Expect(warnings).To(Equal(Warnings{"some-warning"}))
			Expect(jobURL).To(Equal(JobURL("some-job-url")))
		})
	})
})
