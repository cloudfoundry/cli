package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
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

	Describe("CreateOrgRoleByUserName", func() {
		var (
			role       Role
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			role, warnings, executeErr = actor.CreateOrgRoleByUserName(constant.OrgAuditorRole, "user-name", "origin", "org-guid")
		})

		When("the API layer calls are successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRoleReturns(
					ccv3.Role{
						Type:     constant.OrgAuditorRole,
						UserGUID: "user-guid",
						OrgGUID:  "org-guid",
					},
					ccv3.Warnings{"create-role-warning"},
					nil,
				)
			})

			It("returns the role and any warnings", func() {
				Expect(role).To(Equal(
					Role{
						Type:     constant.OrgAuditorRole,
						UserGUID: "user-guid",
						OrgGUID:  "org-guid",
					},
				))
				Expect(warnings).To(ConsistOf("create-role-warning"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(1))
				passedRole := fakeCloudControllerClient.CreateRoleArgsForCall(0)

				Expect(passedRole).To(Equal(
					ccv3.Role{
						Type:     constant.OrgAuditorRole,
						UserName: "user-name",
						Origin:   "origin",
						OrgGUID:  "org-guid",
					},
				))
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

	Describe("CreateSpaceRoleByUserName", func() {
		var (
			returnedRole Role
			warnings     Warnings
			executeErr   error
		)

		JustBeforeEach(func() {
			returnedRole, warnings, executeErr = actor.CreateSpaceRoleByUserName(constant.SpaceDeveloperRole, "user-name", "origin", "space-guid")
		})

		When("the API call to create the returnedRole is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRoleReturns(
					ccv3.Role{
						Type:      constant.SpaceDeveloperRole,
						UserGUID:  "user-guid",
						SpaceGUID: "space-guid",
					},
					ccv3.Warnings{"create-returnedRole-warning"},
					nil,
				)
			})

			It("returns the returnedRole and any warnings", func() {
				Expect(returnedRole).To(Equal(
					Role{
						Type:      constant.SpaceDeveloperRole,
						UserGUID:  "user-guid",
						SpaceGUID: "space-guid",
					},
				))
				Expect(warnings).To(ConsistOf("create-returnedRole-warning"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(1))
				passedRole := fakeCloudControllerClient.CreateRoleArgsForCall(0)

				Expect(passedRole).To(Equal(
					ccv3.Role{
						Type:      constant.SpaceDeveloperRole,
						UserName:  "user-name",
						Origin:    "origin",
						SpaceGUID: "space-guid",
					},
				))
			})
		})

		When("the API call to create the returnedRole returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRoleReturns(
					ccv3.Role{},
					ccv3.Warnings{"create-returnedRole-warning"},
					errors.New("create-returnedRole-error"),
				)
			})

			It("it returns an error and warnings", func() {
				Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(1))
				Expect(warnings).To(ConsistOf("create-returnedRole-warning"))
				Expect(executeErr).To(MatchError("create-returnedRole-error"))
			})
		})
	})

	Describe("CreateOrgRoleByUserGUID", func() {
		var (
			role       Role
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			role, warnings, executeErr = actor.CreateOrgRoleByUserGUID(constant.OrgAuditorRole, "user-guid", "org-guid")
		})

		When("the API layer calls are successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRoleReturns(
					ccv3.Role{
						Type:     constant.OrgAuditorRole,
						UserGUID: "user-guid",
						OrgGUID:  "org-guid",
					},
					ccv3.Warnings{"create-role-warning"},
					nil,
				)
			})

			It("returns the role and any warnings", func() {
				Expect(role).To(Equal(
					Role{
						Type:     constant.OrgAuditorRole,
						UserGUID: "user-guid",
						OrgGUID:  "org-guid",
					},
				))
				Expect(warnings).To(ConsistOf("create-role-warning"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(1))
				passedRole := fakeCloudControllerClient.CreateRoleArgsForCall(0)

				Expect(passedRole).To(Equal(
					ccv3.Role{
						Type:     constant.OrgAuditorRole,
						UserGUID: "user-guid",
						OrgGUID:  "org-guid",
					},
				))
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

	Describe("CreateSpaceRoleByUserGUID", func() {
		var (
			returnedRole Role
			warnings     Warnings
			executeErr   error
		)

		JustBeforeEach(func() {
			returnedRole, warnings, executeErr = actor.CreateSpaceRoleByUserGUID(constant.SpaceDeveloperRole, "user-guid", "space-guid")
		})

		When("the API call to create the returnedRole is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRoleReturns(
					ccv3.Role{
						Type:      constant.SpaceDeveloperRole,
						UserGUID:  "user-guid",
						SpaceGUID: "space-guid",
					},
					ccv3.Warnings{"create-returnedRole-warning"},
					nil,
				)
			})

			It("returns the returnedRole and any warnings", func() {
				Expect(returnedRole).To(Equal(
					Role{
						Type:      constant.SpaceDeveloperRole,
						UserGUID:  "user-guid",
						SpaceGUID: "space-guid",
					},
				))
				Expect(warnings).To(ConsistOf("create-returnedRole-warning"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(1))
				passedRole := fakeCloudControllerClient.CreateRoleArgsForCall(0)

				Expect(passedRole).To(Equal(
					ccv3.Role{
						Type:      constant.SpaceDeveloperRole,
						UserGUID:  "user-guid",
						SpaceGUID: "space-guid",
					},
				))
			})
		})

		When("the API call to create the returnedRole returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRoleReturns(
					ccv3.Role{},
					ccv3.Warnings{"create-returnedRole-warning"},
					errors.New("create-returnedRole-error"),
				)
			})

			It("it returns an error and warnings", func() {
				Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(1))
				Expect(warnings).To(ConsistOf("create-returnedRole-warning"))
				Expect(executeErr).To(MatchError("create-returnedRole-error"))
			})
		})
	})
})
