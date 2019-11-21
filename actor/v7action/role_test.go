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
