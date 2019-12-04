package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Role Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _ = NewTestActor()
	})

	Describe("CreateOrgRole", func() {
		var (
			roleType       constant.RoleType
			userNameOrGUID string
			userOrigin     string
			orgGUID        string
			isClient       bool

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			roleType = constant.OrgAuditorRole
			orgGUID = "org-guid"
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.CreateOrgRole(roleType, orgGUID, userNameOrGUID, userOrigin, isClient)
		})

		When("creating the role succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRoleReturns(
					ccv3.Role{
						Type:     roleType,
						UserGUID: "user-guid",
						OrgGUID:  "org-guid",
					},
					ccv3.Warnings{"create-role-warning"},
					nil,
				)
			})

			When("creating a role for a client", func() {
				BeforeEach(func() {
					userNameOrGUID = "user-guid"
					userOrigin = ""
					isClient = true
				})

				It("returns the role and any warnings", func() {
					Expect(warnings).To(ConsistOf("create-role-warning"))
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(1))
					passedRole := fakeCloudControllerClient.CreateRoleArgsForCall(0)

					Expect(passedRole).To(Equal(
						ccv3.Role{
							Type:     roleType,
							UserGUID: "user-guid",
							OrgGUID:  "org-guid",
						},
					))
				})
			})

			When("creating a role for a non-client user", func() {
				BeforeEach(func() {
					userNameOrGUID = "user-name"
					userOrigin = "user-origin"
					isClient = false
				})

				It("returns the role and any warnings", func() {
					Expect(warnings).To(ConsistOf("create-role-warning"))
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(1))
					passedRole := fakeCloudControllerClient.CreateRoleArgsForCall(0)

					Expect(passedRole).To(Equal(
						ccv3.Role{
							Type:     roleType,
							Username: "user-name",
							Origin:   "user-origin",
							OrgGUID:  "org-guid",
						},
					))
				})
			})
		})

		When("the API call to create the role returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRoleReturns(
					ccv3.Role{},
					ccv3.Warnings{"create-role-warning"},
					errors.New("create-role-error"),
				)
			})

			It("it returns an error and warnings", func() {
				Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(1))
				Expect(warnings).To(ConsistOf("create-role-warning"))
				Expect(executeErr).To(MatchError("create-role-error"))
			})
		})
	})

	Describe("CreateSpaceRole", func() {
		var (
			roleType       constant.RoleType
			userNameOrGUID string
			userOrigin     string
			orgGUID        string
			spaceGUID      string
			isClient       bool

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			roleType = constant.SpaceDeveloperRole
			orgGUID = "org-guid"
			spaceGUID = "space-guid"
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.CreateSpaceRole(roleType, orgGUID, spaceGUID, userNameOrGUID, userOrigin, isClient)
		})

		When("creating the role succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRoleReturnsOnCall(0,
					ccv3.Role{
						Type:     constant.OrgUserRole,
						UserGUID: "user-guid",
						OrgGUID:  "org-guid",
					},
					ccv3.Warnings{"create-org-role-warning"},
					nil,
				)

				fakeCloudControllerClient.CreateRoleReturnsOnCall(1,
					ccv3.Role{
						Type:      roleType,
						UserGUID:  "user-guid",
						SpaceGUID: "space-guid",
					},
					ccv3.Warnings{"create-space-role-warning"},
					nil,
				)
			})

			When("creating a space role for a client", func() {
				BeforeEach(func() {
					userNameOrGUID = "user-guid"
					userOrigin = ""
					isClient = true
				})

				It("returns the role and any warnings", func() {
					Expect(warnings).To(ConsistOf("create-org-role-warning", "create-space-role-warning"))
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(2))

					passedOrgRole := fakeCloudControllerClient.CreateRoleArgsForCall(0)
					Expect(passedOrgRole).To(Equal(
						ccv3.Role{
							Type:     constant.OrgUserRole,
							UserGUID: "user-guid",
							OrgGUID:  "org-guid",
						},
					))

					passedSpaceRole := fakeCloudControllerClient.CreateRoleArgsForCall(1)
					Expect(passedSpaceRole).To(Equal(
						ccv3.Role{
							Type:      roleType,
							UserGUID:  "user-guid",
							SpaceGUID: "space-guid",
						},
					))
				})
			})

			When("creating a space role for a non-client user", func() {
				BeforeEach(func() {
					userNameOrGUID = "user-name"
					userOrigin = "user-origin"
					isClient = false
				})

				It("returns the role and any warnings", func() {
					Expect(warnings).To(ConsistOf("create-org-role-warning", "create-space-role-warning"))
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(2))

					passedOrgRole := fakeCloudControllerClient.CreateRoleArgsForCall(0)
					Expect(passedOrgRole).To(Equal(
						ccv3.Role{
							Type:     constant.OrgUserRole,
							Username: "user-name",
							Origin:   "user-origin",
							OrgGUID:  "org-guid",
						},
					))

					passedSpaceRole := fakeCloudControllerClient.CreateRoleArgsForCall(1)
					Expect(passedSpaceRole).To(Equal(
						ccv3.Role{
							Type:      roleType,
							Username:  "user-name",
							Origin:    "user-origin",
							SpaceGUID: "space-guid",
						},
					))
				})
			})
		})

		When("the user already has an org role", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRoleReturnsOnCall(0,
					ccv3.Role{},
					ccv3.Warnings{"create-org-role-warning"},
					ccerror.RoleAlreadyExistsError{},
				)
			})

			It("it ignores the error and creates the space role", func() {
				Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(2))
				Expect(warnings).To(ConsistOf("create-org-role-warning"))
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})

		When("the API call to create the org role returns an error", func() {
			Context("and it is not a forbidden", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateRoleReturnsOnCall(0,
						ccv3.Role{},
						ccv3.Warnings{"create-org-role-warning"},
						errors.New("create-org-role-error"),
					)
				})

				It("it returns an error and warnings", func() {
					Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(1))
					Expect(warnings).To(ConsistOf("create-org-role-warning"))
					Expect(executeErr).To(MatchError("create-org-role-error"))
				})
			})

			Context("and it is a forbidden", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.CreateRoleReturnsOnCall(0,
						ccv3.Role{},
						ccv3.Warnings{"create-org-role-warning"},
						ccerror.ForbiddenError{Message: "create-org-role-forbidden-error"},
					)
					fakeCloudControllerClient.CreateRoleReturnsOnCall(1,
						ccv3.Role{},
						ccv3.Warnings{"create-space-role-warning"},
						ccerror.ForbiddenError{Message: "create-space-role-forbidden-error"},
					)
				})

				It("it continues to make the call to API create space and returns the more helpful errors and warnings", func() {
					Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(2))
					Expect(warnings).To(ConsistOf("create-space-role-warning", "create-org-role-warning"))
					Expect(executeErr).To(MatchError("create-space-role-forbidden-error"))
				})
			})
		})

		When("the API call to create the space role returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRoleReturnsOnCall(1,
					ccv3.Role{},
					ccv3.Warnings{"create-space-role-warning"},
					errors.New("create-space-role-error"),
				)
			})

			It("it returns an error and warnings", func() {
				Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(2))
				Expect(warnings).To(ConsistOf("create-space-role-warning"))
				Expect(executeErr).To(MatchError("create-space-role-error"))
			})
		})
	})

	Describe("DeleteSpaceRole", func() {
		var (
			roleType       constant.RoleType
			userNameOrGUID string
			userOrigin     string
			spaceGUID      string
			isClient       bool

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			roleType = constant.SpaceDeveloperRole
			spaceGUID = "space-guid"
			isClient = false
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.DeleteSpaceRole(roleType, spaceGUID, userNameOrGUID, userOrigin, isClient)
		})

		When("deleting a role succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetUsersReturnsOnCall(0,
					[]ccv3.User{{Username: userNameOrGUID, GUID: "user-guid"}},
					ccv3.Warnings{"get-users-warning"},
					nil,
				)

				fakeCloudControllerClient.GetUserReturnsOnCall(0,
					ccv3.User{GUID: "user-guid"},
					ccv3.Warnings{"get-user-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRolesReturnsOnCall(0,
					[]ccv3.Role{
						{
							GUID:      "role-guid",
							Type:      roleType,
							UserGUID:  "user-guid",
							SpaceGUID: spaceGUID,
						},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-roles-warning"},
					nil,
				)

				fakeCloudControllerClient.DeleteRoleReturnsOnCall(0,
					ccv3.JobURL("https://jobs/job_guid"),
					ccv3.Warnings{"delete-role-warning"},
					nil,
				)

				fakeCloudControllerClient.PollJobReturnsOnCall(0,
					ccv3.Warnings{"poll-job-warning"},
					nil,
				)

			})

			When("deleting a space role for a client", func() {
				BeforeEach(func() {
					userNameOrGUID = "user-guid"
					userOrigin = ""
					isClient = true
				})

				It("delete the role and returns any warnings", func() {
					Expect(warnings).To(ConsistOf("get-user-warning", "get-roles-warning", "delete-role-warning", "poll-job-warning"))
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.GetUserCallCount()).To(Equal(1))

					passedGuid := fakeCloudControllerClient.GetUserArgsForCall(0)
					Expect(passedGuid).To(Equal(userNameOrGUID))

					passedRolesQuery := fakeCloudControllerClient.GetRolesArgsForCall(0)
					Expect(passedRolesQuery).To(Equal(
						[]ccv3.Query{
							{
								Key:    ccv3.UserGUIDFilter,
								Values: []string{userNameOrGUID},
							},
							{
								Key:    ccv3.RoleTypesFilter,
								Values: []string{string(constant.SpaceDeveloperRole)},
							},
							{
								Key:    ccv3.SpaceGUIDFilter,
								Values: []string{spaceGUID},
							},
						},
					))
					passedRoleGUID := fakeCloudControllerClient.DeleteRoleArgsForCall(0)
					Expect(passedRoleGUID).To(Equal("role-guid"))

					passedJobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
					Expect(passedJobURL).To(Equal(ccv3.JobURL("https://jobs/job_guid")))
				})

			})

			When("deleting a space role for a non-client user", func() {
				BeforeEach(func() {
					userNameOrGUID = "user-name"
					userOrigin = "user-origin"
					isClient = false
				})

				It("deletes the role and returns any warnings", func() {
					Expect(warnings).To(ConsistOf("get-users-warning", "get-roles-warning", "delete-role-warning", "poll-job-warning"))
					Expect(executeErr).ToNot(HaveOccurred())

					passedQuery := fakeCloudControllerClient.GetUsersArgsForCall(0)
					Expect(passedQuery).To(Equal(
						[]ccv3.Query{
							{
								Key:    ccv3.UsernamesFilter,
								Values: []string{userNameOrGUID},
							},
							{
								Key:    ccv3.OriginsFilter,
								Values: []string{userOrigin},
							},
						},
					))

					passedRolesQuery := fakeCloudControllerClient.GetRolesArgsForCall(0)
					Expect(passedRolesQuery).To(Equal(
						[]ccv3.Query{
							{
								Key:    ccv3.UserGUIDFilter,
								Values: []string{"user-guid"},
							},
							{
								Key:    ccv3.RoleTypesFilter,
								Values: []string{string(constant.SpaceDeveloperRole)},
							},
							{
								Key:    ccv3.SpaceGUIDFilter,
								Values: []string{spaceGUID},
							},
						},
					))

					passedRoleGUID := fakeCloudControllerClient.DeleteRoleArgsForCall(0)
					Expect(passedRoleGUID).To(Equal("role-guid"))

					passedJobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
					Expect(passedJobURL).To(Equal(ccv3.JobURL("https://jobs/job_guid")))
				})
			})
		})

		When("the user does not have the space role to delete", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetUsersReturnsOnCall(0,
					[]ccv3.User{{Username: userNameOrGUID, GUID: "user-guid"}},
					ccv3.Warnings{"get-users-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRolesReturnsOnCall(0,
					[]ccv3.Role{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-roles-warning"},
					nil,
				)

			})

			It("it gets an empty list of roles and exits after the request", func() {
				Expect(fakeCloudControllerClient.GetUsersCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.DeleteRoleCallCount()).To(Equal(0))
				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(0))

				Expect(warnings).To(ConsistOf("get-users-warning", "get-roles-warning"))
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})

		Context("the user is not found", func() {
			When("The user is not a client", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetUsersReturnsOnCall(0,
						[]ccv3.User{},
						ccv3.Warnings{"get-users-warning"},
						nil,
					)
				})
				It("returns a user not found error and warnings", func() {
					Expect(fakeCloudControllerClient.GetUsersCallCount()).To(Equal(1))
					passedQuery := fakeCloudControllerClient.GetUsersArgsForCall(0)
					Expect(passedQuery).To(Equal(
						[]ccv3.Query{
							{
								Key:    ccv3.UsernamesFilter,
								Values: []string{userNameOrGUID},
							},
							{
								Key:    ccv3.OriginsFilter,
								Values: []string{userOrigin},
							},
						},
					))
					Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.DeleteRoleCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(0))

					Expect(executeErr).To(MatchError(ccerror.UserNotFoundError{Username: userNameOrGUID, Origin: userOrigin}))
				})
			})

			When("The user is a client", func() {
				BeforeEach(func() {
					userNameOrGUID = "user-guid"
					userOrigin = ""
					isClient = true
					fakeCloudControllerClient.GetUserReturnsOnCall(0,
						ccv3.User{},
						ccv3.Warnings{"get-users-warning"},
						ccerror.UserNotFoundError{},
					)
				})

				It("returns a user not found error and warnings", func() {
					Expect(fakeCloudControllerClient.GetUserCallCount()).To(Equal(1))
					guid := fakeCloudControllerClient.GetUserArgsForCall(0)
					Expect(guid).To(Equal(userNameOrGUID))
					Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.DeleteRoleCallCount()).To(Equal(0))
					Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(0))

					Expect(executeErr).To(MatchError(ccerror.UserNotFoundError{Username: userNameOrGUID}))
				})
			})
		})

		When("the API call to delete the space role returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetUsersReturnsOnCall(0,
					[]ccv3.User{{Username: userNameOrGUID, GUID: "user-guid"}},
					ccv3.Warnings{"get-users-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRolesReturnsOnCall(0,
					[]ccv3.Role{
						{
							GUID:      "role-guid",
							Type:      roleType,
							UserGUID:  "user-guid",
							SpaceGUID: spaceGUID,
						},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-roles-warning"},
					nil,
				)

				fakeCloudControllerClient.DeleteRoleReturnsOnCall(0,
					ccv3.JobURL(""),
					ccv3.Warnings{"delete-space-role-warning"},
					errors.New("delete-space-role-error"),
				)
			})

			It("it returns an error and warnings", func() {
				Expect(fakeCloudControllerClient.DeleteRoleCallCount()).To(Equal(1))
				Expect(executeErr).To(MatchError("delete-space-role-error"))
				Expect(warnings).To(ConsistOf("get-users-warning", "get-roles-warning", "delete-space-role-warning"))
			})
		})
	})

	Describe("GetRoleGUID", func() {
		var (
			roleType  constant.RoleType
			userGUID  string
			spaceGUID string
			roleGUID  string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			roleType = constant.SpaceDeveloperRole
			spaceGUID = "space-guid"
			userGUID = "user-guid"
		})

		JustBeforeEach(func() {
			roleGUID, warnings, executeErr = actor.GetRoleGUID(spaceGUID, userGUID, roleType)
		})
		When("the role exists and no errors occur", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRolesReturnsOnCall(0,
					[]ccv3.Role{
						{
							GUID:      "role-guid",
							Type:      roleType,
							UserGUID:  "user-guid",
							SpaceGUID: spaceGUID,
						},
					},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-roles-warning"},
					nil,
				)
			})
			It("it gets role guid and no errors", func() {
				passedRolesQuery := fakeCloudControllerClient.GetRolesArgsForCall(0)
				Expect(passedRolesQuery).To(Equal(
					[]ccv3.Query{
						{
							Key:    ccv3.UserGUIDFilter,
							Values: []string{"user-guid"},
						},
						{
							Key:    ccv3.RoleTypesFilter,
							Values: []string{string(constant.SpaceDeveloperRole)},
						},
						{
							Key:    ccv3.SpaceGUIDFilter,
							Values: []string{spaceGUID},
						},
					},
				))
				Expect(warnings).To(ConsistOf("get-roles-warning"))
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(roleGUID).To(Equal("role-guid"))
			})
		})

		When("the role does not exist and no errors occur", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRolesReturnsOnCall(0,
					[]ccv3.Role{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-roles-warning"},
					nil,
				)
			})
			It("it gets role guid and no errors", func() {
				passedRolesQuery := fakeCloudControllerClient.GetRolesArgsForCall(0)
				Expect(passedRolesQuery).To(Equal(
					[]ccv3.Query{
						{
							Key:    ccv3.UserGUIDFilter,
							Values: []string{"user-guid"},
						},
						{
							Key:    ccv3.RoleTypesFilter,
							Values: []string{string(constant.SpaceDeveloperRole)},
						},
						{
							Key:    ccv3.SpaceGUIDFilter,
							Values: []string{spaceGUID},
						},
					},
				))
				Expect(warnings).To(ConsistOf("get-roles-warning"))
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(roleGUID).To(Equal(""))
			})
		})

		When("the cloudcontroller returns an error", func() {
			BeforeEach(func() {
				apiError := errors.New("api-get-roles-error")
				fakeCloudControllerClient.GetRolesReturnsOnCall(0,
					[]ccv3.Role{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"get-roles-warning"},
					apiError,
				)
			})
			It("it gets role guid and no errors", func() {
				Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(1))

				Expect(executeErr).To(MatchError("api-get-roles-error"))
			})
		})
	})

	Describe("GetOrgUsersByRoleType", func() {
		var (
			usersByType map[constant.RoleType][]User
			actualErr   error
			warnings    Warnings
		)

		JustBeforeEach(func() {
			usersByType, warnings, actualErr = actor.GetOrgUsersByRoleType("some-org-guid")
		})

		When("when the API returns a success response", func() {
			When("the API returns 2 users", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRolesReturns(
						[]ccv3.Role{
							{
								GUID:     "multiple-user-roleGUID",
								OrgGUID:  "some-org-guid",
								UserGUID: "multi-role-userGUID",
								Type:     constant.OrgUserRole,
							},
							{
								GUID:     "multiple-user-manager-roleGUID",
								OrgGUID:  "some-org-guid",
								UserGUID: "multi-role-userGUID",
								Type:     constant.OrgManagerRole,
							},
							{
								GUID:     "single-user-roleGUID",
								OrgGUID:  "some-org-guid",
								UserGUID: "single-role-userGUID",
								Type:     constant.OrgUserRole,
							},
						},
						ccv3.IncludedResources{
							Users: []ccv3.User{
								{
									Origin:   "uaa",
									Username: "i-have-many-roles",
									GUID:     "multi-role-userGUID",
								},
								{
									Origin:   "uaa",
									Username: "i-have-one-role",
									GUID:     "single-role-userGUID",
								},
							},
						},
						ccv3.Warnings{"some-api-warning"},
						nil,
					)
				})

				It("returns the 2 users", func() {
					Expect(actualErr).NotTo(HaveOccurred())
					Expect(usersByType[constant.OrgUserRole]).To(Equal([]User{
						{
							Origin:   "uaa",
							Username: "i-have-many-roles",
							GUID:     "multi-role-userGUID",
						},
						{
							Origin:   "uaa",
							Username: "i-have-one-role",
							GUID:     "single-role-userGUID",
						},
					}))

					Expect(usersByType[constant.OrgManagerRole]).To(Equal([]User{
						{
							Origin:   "uaa",
							Username: "i-have-many-roles",
							GUID:     "multi-role-userGUID",
						},
					}))

					Expect(warnings).To(ConsistOf("some-api-warning"))

					Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(1))
					query := fakeCloudControllerClient.GetRolesArgsForCall(0)
					Expect(query[0]).To(Equal(ccv3.Query{
						Key:    ccv3.OrganizationGUIDFilter,
						Values: []string{"some-org-guid"},
					}))
					Expect(query[1]).To(Equal(ccv3.Query{
						Key:    ccv3.Include,
						Values: []string{"user"},
					}))
				})
			})
		})

		When("the API returns an error", func() {
			var apiError error

			BeforeEach(func() {
				apiError = errors.New("api-get-roles-error")
				fakeCloudControllerClient.GetRolesReturns(
					[]ccv3.Role{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-warning"},
					apiError,
				)
			})

			It("returns error coming from the API", func() {
				Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(1))

				Expect(actualErr).To(MatchError("api-get-roles-error"))
				Expect(warnings).To(ConsistOf("some-warning"))

			})
		})
	})

	Describe("GetSpaceUsersByRoleType", func() {
		var (
			usersByType map[constant.RoleType][]User
			actualErr   error
			warnings    Warnings
		)

		JustBeforeEach(func() {
			usersByType, warnings, actualErr = actor.GetSpaceUsersByRoleType("some-space-guid")
		})

		When("when the API returns a success response", func() {
			When("the API returns 2 users", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRolesReturns(
						[]ccv3.Role{
							{
								GUID:     "multiple-user-roleGUID",
								UserGUID: "multi-role-userGUID",
								Type:     constant.SpaceDeveloperRole,
							},
							{
								GUID:     "multiple-user-manager-roleGUID",
								UserGUID: "multi-role-userGUID",
								Type:     constant.SpaceManagerRole,
							},
							{
								GUID:     "single-user-roleGUID",
								UserGUID: "single-role-userGUID",
								Type:     constant.SpaceDeveloperRole,
							},
						},
						ccv3.IncludedResources{
							Users: []ccv3.User{
								{
									Origin:   "uaa",
									Username: "i-have-many-roles",
									GUID:     "multi-role-userGUID",
								},
								{
									Origin:   "uaa",
									Username: "i-have-one-role",
									GUID:     "single-role-userGUID",
								},
							},
						},
						ccv3.Warnings{"some-api-warning"},
						nil,
					)
				})

				It("returns the 2 users", func() {
					Expect(actualErr).NotTo(HaveOccurred())
					Expect(usersByType[constant.SpaceDeveloperRole]).To(Equal([]User{
						{
							Origin:   "uaa",
							Username: "i-have-many-roles",
							GUID:     "multi-role-userGUID",
						},
						{
							Origin:   "uaa",
							Username: "i-have-one-role",
							GUID:     "single-role-userGUID",
						},
					}))

					Expect(usersByType[constant.SpaceManagerRole]).To(Equal([]User{
						{
							Origin:   "uaa",
							Username: "i-have-many-roles",
							GUID:     "multi-role-userGUID",
						},
					}))

					Expect(warnings).To(ConsistOf("some-api-warning"))

					Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(1))
					query := fakeCloudControllerClient.GetRolesArgsForCall(0)
					Expect(query[0]).To(Equal(ccv3.Query{
						Key:    ccv3.SpaceGUIDFilter,
						Values: []string{"some-space-guid"},
					}))
					Expect(query[1]).To(Equal(ccv3.Query{
						Key:    ccv3.Include,
						Values: []string{"user"},
					}))
				})
			})
		})

		When("the API returns an error", func() {
			var apiError error

			BeforeEach(func() {
				apiError = errors.New("api-get-roles-error")
				fakeCloudControllerClient.GetRolesReturns(
					[]ccv3.Role{},
					ccv3.IncludedResources{},
					ccv3.Warnings{"some-warning"},
					apiError,
				)
			})

			It("returns error coming from the API", func() {
				Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(1))

				Expect(actualErr).To(MatchError("api-get-roles-error"))
				Expect(warnings).To(ConsistOf("some-warning"))

			})
		})
	})
})
