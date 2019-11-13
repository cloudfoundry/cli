package v7action_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
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
							UserName: "user-name",
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
							UserName: "user-name",
							Origin:   "user-origin",
							OrgGUID:  "org-guid",
						},
					))

					passedSpaceRole := fakeCloudControllerClient.CreateRoleArgsForCall(1)
					Expect(passedSpaceRole).To(Equal(
						ccv3.Role{
							Type:      roleType,
							UserName:  "user-name",
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

	Describe("GetOrgRole", func() {
		var (
			roleType = constant.OrgAuditorRole
			orgGUID  = "org-guid"
			userGUID = "user-guid"

			executeErr error
			warnings   Warnings
			role       Role
		)

		JustBeforeEach(func() {
			role, warnings, executeErr = actor.GetOrgRole(roleType, orgGUID, userGUID)
		})

		When("the cc client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRolesReturns(
					nil,
					ccv3.Warnings{"get-roles-warning"},
					errors.New("get-roles-error"),
				)
			})

			It("returns warnings and the error", func() {
				Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(1))
				actualQueries := fakeCloudControllerClient.GetRolesArgsForCall(0)
				Expect(actualQueries).To(ConsistOf(
					ccv3.Query{Key: ccv3.TypeFilter, Values: []string{string(roleType)}},
					ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
					ccv3.Query{Key: ccv3.UserGUIDFilter, Values: []string{userGUID}},
				))

				Expect(warnings).To(ConsistOf("get-roles-warning"))
				Expect(executeErr).To(MatchError(errors.New("get-roles-error")))
			})
		})

		When("the cc client succeeds and a route is found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRolesReturns(
					[]ccv3.Role{{GUID: "role-guid"}},
					ccv3.Warnings{"get-roles-warning"},
					nil,
				)
			})

			It("returns the route and the warnings", func() {
				Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(1))
				actualQueries := fakeCloudControllerClient.GetRolesArgsForCall(0)
				Expect(actualQueries).To(ConsistOf(
					ccv3.Query{Key: ccv3.TypeFilter, Values: []string{string(roleType)}},
					ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
					ccv3.Query{Key: ccv3.UserGUIDFilter, Values: []string{userGUID}},
				))

				Expect(warnings).To(ConsistOf("get-roles-warning"))
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(role).To(Equal(Role{GUID: "role-guid"}))
			})
		})

		When("the cc client succeeds and a route is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRolesReturns([]ccv3.Role{}, ccv3.Warnings{"get-roles-warning"}, nil)
			})

			It("returns the role and the warnings", func() {
				Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(1))
				actualQueries := fakeCloudControllerClient.GetRolesArgsForCall(0)
				Expect(actualQueries).To(ConsistOf(
					ccv3.Query{Key: ccv3.TypeFilter, Values: []string{string(roleType)}},
					ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
					ccv3.Query{Key: ccv3.UserGUIDFilter, Values: []string{userGUID}},
				))

				Expect(warnings).To(ConsistOf("get-roles-warning"))
				Expect(executeErr).To(MatchError(actionerror.RoleNotFoundError{}))
			})
		})
	})

	Describe("GetSpaceRole", func() {
		var (
			roleType  = constant.SpaceAuditorRole
			spaceGUID = "space-guid"
			userGUID  = "user-guid"

			executeErr error
			warnings   Warnings
			role       Role
		)

		JustBeforeEach(func() {
			role, warnings, executeErr = actor.GetSpaceRole(roleType, spaceGUID, userGUID)
		})

		When("the cc client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRolesReturns(
					nil,
					ccv3.Warnings{"get-roles-warning"},
					errors.New("get-roles-error"),
				)
			})

			It("returns warnings and the error", func() {
				Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(1))
				actualQueries := fakeCloudControllerClient.GetRolesArgsForCall(0)
				Expect(actualQueries).To(ConsistOf(
					ccv3.Query{Key: ccv3.TypeFilter, Values: []string{string(roleType)}},
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
					ccv3.Query{Key: ccv3.UserGUIDFilter, Values: []string{userGUID}},
				))

				Expect(warnings).To(ConsistOf("get-roles-warning"))
				Expect(executeErr).To(MatchError(errors.New("get-roles-error")))
			})
		})

		When("the cc client succeeds and a route is found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRolesReturns(
					[]ccv3.Role{{GUID: "role-guid"}},
					ccv3.Warnings{"get-roles-warning"},
					nil,
				)
			})

			It("returns the route and the warnings", func() {
				Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(1))
				actualQueries := fakeCloudControllerClient.GetRolesArgsForCall(0)
				Expect(actualQueries).To(ConsistOf(
					ccv3.Query{Key: ccv3.TypeFilter, Values: []string{string(roleType)}},
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
					ccv3.Query{Key: ccv3.UserGUIDFilter, Values: []string{userGUID}},
				))

				Expect(warnings).To(ConsistOf("get-roles-warning"))
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(role).To(Equal(Role{GUID: "role-guid"}))
			})
		})

		When("the cc client succeeds and a route is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRolesReturns([]ccv3.Role{}, ccv3.Warnings{"get-roles-warning"}, nil)
			})

			It("returns the role and the warnings", func() {
				Expect(fakeCloudControllerClient.GetRolesCallCount()).To(Equal(1))
				actualQueries := fakeCloudControllerClient.GetRolesArgsForCall(0)
				Expect(actualQueries).To(ConsistOf(
					ccv3.Query{Key: ccv3.TypeFilter, Values: []string{string(roleType)}},
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
					ccv3.Query{Key: ccv3.UserGUIDFilter, Values: []string{userGUID}},
				))

				Expect(warnings).To(ConsistOf("get-roles-warning"))
				Expect(executeErr).To(MatchError(actionerror.RoleNotFoundError{}))
			})
		})
	})
})
